package authentication // import "breve.us/authsvc/authentication"

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"

	"breve.us/authsvc/common"
	"breve.us/authsvc/oauth"
)

// Exported Errors
var (
	ErrInvalidRequest    = errors.New("invalid_request")
	ErrInvalidToken      = errors.New("invalid_token")
	ErrInsufficientScope = errors.New("insufficient_scope")
	ErrNoAuthToken       = errors.New("no_auth")
)

var (
	loginPath     = "login/"
	logoutPath    = "logout/"
	cookieName    = "authsvc-login-cookie"
	redirectParam = "redirect_uri"
	loginLifetime = 60 * 60 * 2 // 2 hours

	messages = map[string]string{
		ErrInvalidRequest.Error():    "malformed request",
		ErrInvalidToken.Error():      "invalid or expired authentication token",
		ErrInsufficientScope.Error(): "insufficient scope granted",
	}

	errorPages = map[string]string{
		ErrInvalidRequest.Error(): loginPath,
		ErrInvalidToken.Error():   loginPath,
	}
)

// Options provides configuration options to the AuthenticationMiddleware.
type Options struct {
	Realm       string
	Seeder      common.Seeder
	PublicRoots []string
	OAuth       *oauth.Handler
	Insecure    bool
}

// NewMiddleware returns a middlware suitable for authentication.
func NewMiddleware(root string, config *Options) (*Middleware, error) {
	var err error
	if config == nil {
		config = &Options{}
	}
	if config.Seeder == nil {
		if config.Seeder, err = common.NewDefaultSeeder(); err != nil {
			return nil, err
		}
	}
	if config.OAuth == nil {
		config.OAuth = oauth.NewHandler(nil)
	}
	return &Middleware{
		authRoot: root,
		config:   config,
		cookie:   securecookie.New(config.Seeder.HashKey(), config.Seeder.BlockKey()),
	}, nil
}

// Middleware enforces authentication on protected routes.
type Middleware struct {
	authRoot string
	config   *Options
	cookie   *securecookie.SecureCookie
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
	if m.whitelisted(r) {
		n(w, r)
		return
	}

	switch err := m.authenticated(r); err {
	case nil:
		n(w, r)
	case ErrNoAuthToken:
		m.unauthorized(w, r.URL, nil)
	default:
		m.clearLoginCookie(w)
		m.unauthorized(w, r.URL, err)
	}
}

// LoginHandler returns a router that handles the login and logout routes.
func (m *Middleware) LoginHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc(m.authRoot+loginPath, m.loginPOST).Methods("POST")
	r.HandleFunc(m.authRoot+logoutPath, m.logoutPOST).Methods("POST")
	return r
}

func (m *Middleware) loginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch button := r.Form.Get("submit"); button {
	case "Login":
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if !m.validate(username, password) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.setLoginCookie(username, w)
		returnURL, err := url.PathUnescape(r.Form.Get(redirectParam))
		if err != nil {
			returnURL = "/"
		}
		common.Redirect(w, r, returnURL, nil)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *Middleware) logoutPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch button := r.Form.Get("submit"); button {
	case "Logout":
		m.clearLoginCookie(w)
		common.Redirect(w, r, "/", map[string]string{"msg": "logged out"})
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *Middleware) setLoginCookie(username string, w http.ResponseWriter) {
	data := map[string]string{"username": username}
	switch v, err := m.cookie.Encode(cookieName, data); err {
	case nil:
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    v,
			Path:     "/",
			Expires:  time.Now().Add(time.Duration(loginLifetime) * time.Second),
			MaxAge:   loginLifetime,
			Secure:   !m.config.Insecure,
			HttpOnly: true,
		})
	default:
		log.Printf("error encoding the cookie %q with username %q: %v", cookieName, username, err)
		common.JSONStatusResponse(http.StatusInternalServerError, w, "error logging in")
	}
}
func (m *Middleware) clearLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func (m *Middleware) validate(username, password string) bool {
	if username == "" || password == "" {
		return false
	}
	if username == "jweldon" && password == "password" {
		return true
	}
	return false
}

func (m *Middleware) whitelisted(r *http.Request) bool {
	path := r.URL.Path
	for _, root := range m.config.PublicRoots {
		if strings.HasPrefix(path, root) {
			return true
		}
	}
	return false
}

func (m *Middleware) authenticated(r *http.Request) error {
	if r.Header.Get("Authorization") != "" {
		if scopes, err := m.config.OAuth.Authorized(r); err == nil && len(scopes) > 0 {
			// TODO: validate scopes
			return nil
		}
	}
	switch c, err := r.Cookie(cookieName); err {
	case nil:
		data := map[string]string{}
		if err := m.cookie.Decode(cookieName, c.Value, &data); err != nil {
			return ErrInvalidRequest
		}
		// TODO: verify username (locked out, etc)
		return nil
	case http.ErrNoCookie:
		return ErrNoAuthToken
	default:
		return ErrInvalidRequest
	}
}

func (m *Middleware) unauthorized(w http.ResponseWriter, originalURL fmt.Stringer, err error) {
	var b strings.Builder
	_, _ = b.WriteString("authsvc")

	code := http.StatusUnauthorized
	parts := []string{}
	if m.config.Realm != "" {
		parts = append(parts, fmt.Sprintf("realm=%q", m.config.Realm))
	}
	switch err {
	case nil, ErrNoAuthToken:
		// No error message
	default:
		errmsg := err.Error()
		parts = append(parts, fmt.Sprintf("error=%q", errmsg))
		if msg, ok := messages[errmsg]; ok {
			parts = append(parts, fmt.Sprintf("error_message=%q", msg))
		}
		if page, ok := errorPages[errmsg]; ok {
			parts = append(parts, fmt.Sprintf("error_uri=%q", m.authRoot+page))
		}
	}

	if len(parts) > 0 {
		_, _ = b.WriteString(" " + strings.Join(parts, ", "))
	}

	login := makeRedirect(m.authRoot+loginPath, originalURL)
	w.WriteHeader(code)
	w.Header().Set("WWW-Authenticate", b.String())
	w.Header().Set("Location", login.String())

	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="0;URL='%s'" /></head><body></body></html>`, login.String())
}

func makeRedirect(target string, originalURL fmt.Stringer) fmt.Stringer {
	var (
		u   *url.URL
		err error
	)
	if u, err = url.Parse(target); err != nil {
		return originalURL
	}
	v := u.Query()
	v.Set(redirectParam, url.QueryEscape(originalURL.String()))
	u.RawQuery = v.Encode()
	return u
}
