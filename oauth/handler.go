package oauth // import "breve.us/authsvc/oauth"

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/gorilla/mux"

	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
)

func init() {
	rand.Seed(time.Now().Unix())
	gob.Register(&authorize{})
}

// Error Values
var (
	ErrNotAuthorized       = errors.New("not authorized")
	ErrInvalidClient       = errors.New("invalid client")
	ErrInvalidAuth         = errors.New("invalid auth")
	ErrMissingRedirect     = errors.New("missing redirect_uri")
	ErrInvalidResponseType = errors.New("response_type unsupported")
)

// Scopes
const (
	ScopeAll = "all"
)

// Options encapsulates OAuth Handler options.
type Options struct {
	TokenTTL   time.Duration
	GrantTTL   time.Duration
	Cache      store.Cache
	TokenCache *TokenCache
	Clients    *ClientRegistry
}

// RegisterAPI returns a router that handles OAuth routes.
func (h *Handler) RegisterAPI(root string) http.Handler {
	mx := mux.NewRouter()
	mx.Path(path.Join(root, "authorize")).HandlerFunc(h.handleAuthorize).Methods("GET")
	mx.Path(path.Join(root, "approve")).HandlerFunc(h.handleApprove).Methods("GET")
	mx.Path(path.Join(root, "token")).HandlerFunc(h.handleToken).Methods("POST")
	mx.PathPrefix(root).HandlerFunc(h.handleDefault).Methods("GET")
	return mx
}

// NewHandler creates and initializes a new OAuthHandler
func NewHandler(options *Options) *Handler {
	if options == nil {
		options = &Options{}
	}
	if options.TokenTTL == 0 {
		options.TokenTTL = 15 * time.Minute
	}
	if options.GrantTTL == 0 {
		options.GrantTTL = 14 * 24 * time.Hour
	}
	if options.Cache == nil {
		options.Cache = store.NewMemoryCache()
	}
	if options.TokenCache == nil {
		options.TokenCache = NewTokenCache(store.NewMemoryCache(), store.NewMemoryCache())
	}
	return &Handler{opts: options}
}

// Handler provides OAuth2 capabilities.
type Handler struct {
	opts *Options
}

// Authorized returns the authorized scopes for a request, or an error
// if the request does not have sufficient authorization.
func (h *Handler) Authorized(r *http.Request) ([]string, error) {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		switch {
		case strings.HasPrefix(auth, "Bearer "):
			if h.validToken(auth[7:]) {
				return []string{ScopeAll}, nil
			}
		}
	}
	return nil, ErrNotAuthorized
}

func (h *Handler) validToken(token string) bool {
	_, err := h.opts.TokenCache.Get(token)
	// TODO: validate client id here (locked out, etc.)
	return err == nil
}

func (h *Handler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	a := parseAuthorize(r.URL.Query())
	if err := a.valid(); err != nil {
		common.JSONStatusResponse(http.StatusForbidden, w, err.Error())
		return
	}
	if !h.opts.Clients.VerifyClient(a.ClientID) {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid client id")
		return
	}
	if !h.opts.Clients.VerifyRedirect(a.ClientID, a.RedirectURI) {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid client redirect")
		return
	}

	h.addToCache(a)
	a.serveForm(w)
}

func (h *Handler) handleApprove(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	a, ok := h.checkCode(r.Form.Get("corr"))
	if !ok {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid correlation")
		return
	}
	if r.Form.Get("approve") != "Approve" {
		common.Redirect(w, r, a.RedirectURI, map[string]string{"error": "access_denied"})
		return
	}
	if a.ResponseType != "code" {
		common.Redirect(w, r, a.RedirectURI, map[string]string{"error": "unsupported_response_type"})
		return
	}
	code := h.addToCache(a)
	common.Redirect(w, r, a.RedirectURI, map[string]string{"code": code, "state": a.State})
}

func (h *Handler) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	t, ok := h.checkCode(r.Form.Get("code"))
	if !ok {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid token")
		return
	}

	creds, err := decodeClientCredentials(r)
	if err != nil {
		common.JSONStatusResponse(http.StatusForbidden, w, err.Error())
		return
	}

	if !h.opts.Clients.VerifyClient(creds.ID) {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid client id")
		return
	}

	switch r.Form.Get("grant_type") {
	case "authorization_code":
		if t.ClientID != creds.ID {
			common.JSONStatusResponse(http.StatusForbidden, w, "mismatching client ids")
			return
		}
		switch err := h.opts.Cache.Delete(t.ID); err {
		case nil, store.ErrNotFound:
		default:
			panic(err)
		}
		tok := generateRandomString()
		h.addClient(tok, creds.ID)
		common.JSONResponse(w, &bearer{Token: tok, Type: "Bearer"})
	default:
		common.JSONStatusResponse(http.StatusForbidden, w, "unsupported grant type")
		return
	}
}

func (h *Handler) handleDefault(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	obj := struct {
		URL       string
		QueryKeys []string
		Values    interface{}
	}{URL: r.URL.String(), QueryKeys: common.QueryParamKeys(r.Form), Values: r.Form}

	common.JSONResponse(w, obj)
}

func (h *Handler) addClient(tok string, id string) {
	expire := time.Now().Add(h.opts.GrantTTL)
	if err := h.opts.TokenCache.PutUntil(expire, id, tok); err != nil {
		panic(err)
	}
}

func (h *Handler) addToCache(a *authorize) string {
	if a.ID == "" {
		a.ID = generateRandomString()
	}
	expire := time.Now().Add(h.opts.TokenTTL)
	if err := h.opts.Cache.PutUntil(expire, a.ID, a); err != nil {
		panic(err)
	}
	return a.ID
}

func (h *Handler) checkCode(code string) (*authorize, bool) {
	switch v, err := h.opts.Cache.Get(code); err {
	case nil:
		a, ok := v.(*authorize)
		return a, ok
	default:
		log.Printf("error checkCode(%q): %v", code, err)
		return nil, false
	}
}

func generateRandomString() string { return fmt.Sprintf("%X", rand.Int63()) }

type bearer struct {
	Token string `json:"access_token"`
	Type  string `json:"token_type"`
}

type clientCredentials struct {
	ID     string
	Secret string
}

func decodeClientCredentials(r *http.Request) (clientCredentials, error) {
	cred := clientCredentials{
		ID:     r.Form.Get("client_id"),
		Secret: r.Form.Get("client_secret"),
	}
	auth := r.Header.Get("Authorization")
	if auth != "" {
		if cred.ID != "" {
			return cred, ErrInvalidClient
		}
		switch {
		case strings.HasPrefix(auth, "Basic "):
			data, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				return cred, err
			}
			a := strings.SplitN(string(data), ":", 2)
			if len(a) < 2 {
				return cred, ErrInvalidAuth
			}
			cred.ID = a[0]
			cred.Secret = a[1]
		}
	}
	return cred, nil
}

type authorize struct {
	ID           string
	Application  string
	ResponseType string
	ClientID     string
	RedirectURI  string
	State        string
}

func parseAuthorize(v url.Values) *authorize {
	return &authorize{
		Application:  "MatterMost",
		ResponseType: v.Get("response_type"),
		ClientID:     v.Get("client_id"),
		RedirectURI:  v.Get("redirect_uri"),
		State:        v.Get("state"),
	}
}

func (a *authorize) valid() error {
	if a.RedirectURI == "" {
		return ErrMissingRedirect
	}
	if _, err := url.Parse(a.RedirectURI); err != nil {
		return err
	}
	if a.ResponseType != "code" {
		return ErrInvalidResponseType
	}
	return nil
}

func (a *authorize) serveForm(w http.ResponseWriter) {
	if err := authorizeForm.Execute(w, a); err != nil {
		common.JSONStatusResponse(http.StatusInternalServerError, w, nil)
	}
}

var authorizeForm = template.Must(template.New("authorize").Parse(`<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <title>Login</title>
  </head>
  <body>
    <h3>Allow {{ .Application }}?</h3>
    <form action="/oauth/approve" method="get">
      <div>
        <input type="hidden" name="corr" value="{{ .ID }}" />
        <input type="submit" name="approve" value="Approve"></input>
        <input type="submit" name="deny" value="Deny"></input>
      </div>
    </form>
  </body>
</html>
`))
