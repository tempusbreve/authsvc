package authsvc // import "breve.us/authsvc"

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/alecthomas/template"
)

func newOAuthHandler() http.Handler {
	return &oauth{
		cache:   map[string]*authorize{},
		clients: map[string]string{},
		tokens:  map[string]string{},
	}
}

type oauth struct {
	cache   map[string]*authorize
	clients map[string]string
	tokens  map[string]string
}

func (h *oauth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONCode(http.StatusBadRequest, w, err.Error())
		return
	}

	switch p := r.URL.Path; p {
	case "/oauth/authorize":
		h.handleAuthorize(w, r)
	case "/oauth/approve":
		h.handleApprove(w, r)
	case "/oauth/token":
		h.handleToken(w, r)
	default:
		h.handleDefault(w, r)
	}
}

func (h *oauth) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
	default:
		writeJSONCode(http.StatusMethodNotAllowed, w, r.Method)
		return
	}

	a := parseAuthorize(r.URL.Query())
	if err := a.valid(); err != nil {
		writeJSONCode(http.StatusForbidden, w, err.Error())
		return
	}
	if !h.checkClient(a.ClientID) {
		writeJSONCode(http.StatusForbidden, w, "invalid client id")
		return
	}
	if !h.checkRedirect(a) {
		writeJSONCode(http.StatusForbidden, w, "invalid redirect uri")
		return
	}
	h.addToCache(a)
	a.serveForm(w)
}

func (h *oauth) handleApprove(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
	default:
		writeJSONCode(http.StatusMethodNotAllowed, w, r.Method)
		return
	}

	a, ok := h.checkCorrelation(r.Form.Get("corr"))
	if !ok {
		writeJSONCode(http.StatusForbidden, w, "")
		return
	}
	if r.Form.Get("approve") != "Approve" {
		writeRedirect(w, a.RedirectURI, map[string]string{"error": "access_denied"})
		return
	}
	if a.ResponseType != "code" {
		writeRedirect(w, a.RedirectURI, map[string]string{"error": "unsupported_response_type"})
		return
	}
	code := h.addToCache(a)
	writeRedirect(w, a.RedirectURI, map[string]string{"code": code, "state": a.State})
}

func (h *oauth) handleToken(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
	default:
		writeJSONCode(http.StatusMethodNotAllowed, w, r.Method)
		return
	}

	t, ok := h.checkToken(r.Form.Get("code"))
	if !ok {
		writeJSONCode(http.StatusForbidden, w, "")
		return
	}

	creds, err := decodeClientCredentials(r)
	if err != nil {
		writeJSONCode(http.StatusForbidden, w, err.Error())
		return
	}

	if !h.checkClient(creds.ID) {
		writeJSONCode(http.StatusForbidden, w, err.Error())
		return
	}

	switch r.Form.Get("grant_type") {
	case "authorization_code":
		if t.ClientID != creds.ID {
			writeJSONCode(http.StatusForbidden, w, "mismatching client ids")
			return
		}
		h.removeFromCache(t.ID)
		tok := generateRandomString()
		h.addClient(tok, creds.ID)
		writeJSON(w, &bearer{Token: tok, Type: "Bearer"})
	default:
		writeJSONCode(http.StatusForbidden, w, "unsupported grant type")
		return
	}
}

func (h *oauth) handleDefault(w http.ResponseWriter, r *http.Request) {
	obj := struct {
		Url       string
		QueryKeys []string
		Values    interface{}
	}{Url: r.URL.String(), QueryKeys: queryKeys(r.Form), Values: r.Form}

	writeJSON(w, obj)
}

func (h *oauth) addClient(tok string, id string) {
	h.clients[id] = tok
	h.tokens[tok] = id
}

func (h *oauth) addToCache(a *authorize) string {
	if a.ID != "" {
		h.removeFromCache(a.ID)
		a.ID = ""
	}
	a.ID = generateRandomString()
	h.cache[a.ID] = a
	return a.ID
}

func (h *oauth) removeFromCache(id string) { delete(h.cache, id) }

func (h *oauth) checkToken(tok string) (*authorize, bool) {
	t, ok := h.cache[tok]
	return t, ok
}

func (h *oauth) checkCorrelation(corr string) (*authorize, bool) {
	// TODO: more validation?
	return h.checkToken(corr)
}

func (h *oauth) checkRedirect(a *authorize) bool {
	// TODO: implement
	return true
}

func (h *oauth) checkClient(id string) bool {
	// TODO: implement
	return true
}

func generateRandomString() string {
	// TODO: fix
	return fmt.Sprintf("%X", rand.Int63())
}

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
			return cred, errors.New("invalid client")
		}
		switch {
		case strings.HasPrefix(auth, "Basic "):
			data, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				return cred, err
			}
			a := strings.SplitN(string(data), ":", 2)
			if len(a) < 2 {
				return cred, errors.New("invalid auth")
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
		return errors.New("missing redirect_uri")
	}
	if _, err := url.Parse(a.RedirectURI); err != nil {
		return err
	}
	if a.ResponseType != "code" {
		return errors.New("response_type unsupported")
	}
	return nil
}

func (a *authorize) serveForm(w http.ResponseWriter) {
	if err := authorizeForm.Execute(w, a); err != nil {
		writeJSONCode(http.StatusInternalServerError, w, nil)
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