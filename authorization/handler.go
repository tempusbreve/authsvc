package authorization // import "breve.us/authsvc/authorization"

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/gorilla/mux"

	"breve.us/authsvc/client"
	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
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
	TokenTTL time.Duration
	GrantTTL time.Duration
	CacheDir string
	Users    *user.Registry
}

// RegisterAPI returns a router that handles OAuth routes.
func (h *OAuthHandler) RegisterAPI(root string) http.Handler {
	mx := mux.NewRouter()
	mx.Path(path.Join(root, "authorize")).HandlerFunc(h.handleAuthorize).Methods("GET")
	mx.Path(path.Join(root, "approve")).HandlerFunc(h.handleApprove).Methods("GET")
	mx.Path(path.Join(root, "token")).HandlerFunc(h.handleToken).Methods("POST")
	mx.PathPrefix(root).HandlerFunc(h.handleDefault).Methods("GET")
	return mx
}

// NewHandler creates and initializes a new OAuthHandler
func NewHandler(options *Options) (*OAuthHandler, error) {
	if options == nil {
		options = &Options{}
	}
	if options.TokenTTL == 0 {
		options.TokenTTL = 15 * time.Minute
	}
	if options.GrantTTL == 0 {
		options.GrantTTL = 14 * 24 * time.Hour
	}

	var (
		cache store.Cache
		tok   *tokenCache
	)

	cr := client.NewRegistry(store.NewMemoryCache())

	if validDir(options.CacheDir) {
		var (
			err    error
			cc, tc store.Cache
			fd     io.ReadCloser
		)
		if cache, err = store.NewBoltDBCache(path.Join(options.CacheDir, "transient.db"), "cache"); err != nil {
			return nil, err
		}
		if cc, err = store.NewBoltDBCache(path.Join(options.CacheDir, "tokens.db"), "clients"); err != nil {
			return nil, err
		}
		if tc, err = store.NewBoltDBCache(path.Join(options.CacheDir, "tokens.db"), "tokens"); err != nil {
			return nil, err
		}
		tok = newTokenCache(cc, tc)

		if fd, err = os.Open(path.Join(options.CacheDir, "clients.json")); err != nil {
			return nil, err
		}
		defer func() { _ = fd.Close() }()
		if err = cr.LoadFromJSON(fd); err != nil {
			return nil, err
		}
	} else {
		cache = store.NewMemoryCache()
		tok = newTokenCache(store.NewMemoryCache(), store.NewMemoryCache())
	}

	return &OAuthHandler{
		opts:    options,
		cache:   cache,
		tokens:  tok,
		clients: cr,
		checker: newTokenRequestChecker(tok, options.Users),
	}, nil
}

// OAuthHandler provides OAuth2 capabilities.
type OAuthHandler struct {
	opts    *Options
	cache   store.Cache
	tokens  *tokenCache
	clients *client.Registry
	checker common.RequestChecker
}

// Authorized returns the authorized scopes for a request, or an error
// if the request does not have sufficient authorization.
func (h *OAuthHandler) Authorized(r *http.Request) ([]string, error) {
	if h.checker.IsAuthenticated(r) != "" {
		return []string{ScopeAll}, nil
	}
	return nil, ErrNotAuthorized
}

func (h *OAuthHandler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	a := parseAuthorize(r.URL.Query())
	if err := a.valid(); err != nil {
		common.JSONStatusResponse(http.StatusForbidden, w, err.Error())
		return
	}
	if !h.clients.VerifyClient(a.ClientID) {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid client id")
		return
	}
	if !h.clients.VerifyRedirect(a.ClientID, a.RedirectURI) {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid client redirect")
		return
	}

	h.addToCache(a)
	a.serveForm(w)
}

func (h *OAuthHandler) handleApprove(w http.ResponseWriter, r *http.Request) {
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
	if username := common.GetUsername(r.Context()); username != "" {
		a.Username = username
	}
	code := h.addToCache(a)
	common.Redirect(w, r, a.RedirectURI, map[string]string{"code": code, "state": a.State})
}

func (h *OAuthHandler) handleToken(w http.ResponseWriter, r *http.Request) {
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

	if !h.clients.VerifyClient(creds.ID) {
		common.JSONStatusResponse(http.StatusForbidden, w, "invalid client id")
		return
	}

	switch r.Form.Get("grant_type") {
	case "authorization_code":
		if t.ClientID != creds.ID {
			common.JSONStatusResponse(http.StatusForbidden, w, "mismatching client ids")
			return
		}
		switch err := h.cache.Delete(t.ID); err {
		case nil, store.ErrNotFound:
		default:
			panic(err)
		}
		tok := generateRandomString()
		h.addClient(tok, t.Username)
		common.JSONResponse(w, &bearer{Token: tok, Type: "Bearer"})
	default:
		common.JSONStatusResponse(http.StatusForbidden, w, "unsupported grant type")
		return
	}
}

func (h *OAuthHandler) handleDefault(w http.ResponseWriter, r *http.Request) {
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

func (h *OAuthHandler) addClient(tok string, username string) {
	expire := time.Now().Add(h.opts.GrantTTL)
	if err := h.tokens.PutUntil(expire, username, tok); err != nil {
		panic(err)
	}
}

func (h *OAuthHandler) addToCache(a *authorize) string {
	if a.ID == "" {
		a.ID = generateRandomString()
	}
	expire := time.Now().Add(h.opts.TokenTTL)
	if err := h.cache.PutUntil(expire, a.ID, a); err != nil {
		panic(err)
	}
	return a.ID
}

func (h *OAuthHandler) checkCode(code string) (*authorize, bool) {
	switch v, err := h.cache.Get(code); err {
	case nil:
		a, ok := v.(*authorize)
		return a, ok
	default:
		log.Printf("error checkCode(%q): %v", code, err)
		return nil, false
	}
}

func validDir(dir string) bool {
	if dir == "" {
		return false
	}
	s, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return s.IsDir()
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
	Username     string
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
