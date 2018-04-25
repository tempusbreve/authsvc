package authsvc // import "breve.us/authsvc"

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
)

var (
	ErrInvalidRequest    = errors.New("invalid_request")
	ErrInvalidToken      = errors.New("invalid_token")
	ErrInsufficientScope = errors.New("insufficient_scope")
	ErrNoAuthToken       = errors.New("no_auth")

	loginPath     = "login/"
	cookieName    = "authsvc-login-cookie"
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

func NewAuthenticationMiddleware(realm string, root string, seeder Seeder) *authenticationMiddleware {
	if seeder == nil {
		seeder = NewSeeder()
	}
	return &authenticationMiddleware{
		realm:    realm,
		authRoot: root,
		cookie:   securecookie.New(seeder.HashKey(), seeder.BlockKey()),
	}
}

type authenticationMiddleware struct {
	realm    string
	authRoot string
	cookie   *securecookie.SecureCookie
}

func (m *authenticationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
	switch err := m.authenticated(r); err {
	case nil:
		n(w, r)
	case ErrNoAuthToken:
		m.unauthorized(w, r, nil)
	default:
		m.clearLoginCookie(w)
		m.unauthorized(w, r, err)
	}
}

func (m *authenticationMiddleware) LoginHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc(m.authRoot+loginPath, m.loginPOST).Methods("POST")
	return r
}

func (m *authenticationMiddleware) loginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch button := r.Form.Get("submit"); button {
	case "Logout":
		m.clearLoginCookie(w)
		writeRedirect(w, r, "/", map[string]string{"msg": "logged out"})
	case "Login":
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if !m.validate(username, password) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.setLoginCookie(username, w)
		writeRedirect(w, r, "/", map[string]string{"msg": "logged in"})
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *authenticationMiddleware) setLoginCookie(username string, w http.ResponseWriter) {
	data := map[string]string{"username": username}
	switch v, err := m.cookie.Encode(cookieName, data); err {
	case nil:
		http.SetCookie(w, &http.Cookie{
			Name:    cookieName,
			Value:   v,
			Path:    "/",
			Expires: time.Now().Add(time.Duration(loginLifetime) * time.Second),
			MaxAge:  loginLifetime,
			/*
				Secure:   true,
				HttpOnly: true,
			*/
		})
	default:
		log.Printf("error encoding the cookie %q with username %q: %v", cookieName, username, err)
		writeJSONCode(http.StatusInternalServerError, w, "error logging in")
	}
}
func (m *authenticationMiddleware) clearLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func (m *authenticationMiddleware) validate(username, password string) bool {
	if username == "" || password == "" {
		return false
	}
	if username == "jweldon" && password == "password" {
		return true
	}
	return false
}

func (m *authenticationMiddleware) authenticated(r *http.Request) error {
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

func (m *authenticationMiddleware) unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	var b strings.Builder
	b.WriteString("authsvc")

	code := http.StatusUnauthorized
	parts := []string{}
	if m.realm != "" {
		parts = append(parts, fmt.Sprintf("realm=%q", m.realm))
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
		b.WriteString(" " + strings.Join(parts, ", "))
	}

	w.WriteHeader(code)
	w.Header().Set("WWW-Authenticate", b.String())
	w.Header().Set("Location", m.authRoot+loginPath)

	login, err := url.Parse(m.authRoot + loginPath + "?return_url=" + url.PathEscape(r.URL.Path))
	if err != nil {
		if login, err = url.Parse(m.authRoot + loginPath); err != nil {
			panic(err)
		}
	}

	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="0;URL='%s'" /></head><body></body></html>`, login.String())
}
