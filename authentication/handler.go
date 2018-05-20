package authentication // import "breve.us/authsvc/authentication"

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"

	"breve.us/authsvc/common"
)

var (
	loginPath     = "login/"
	logoutPath    = "logout/"
	cookieName    = "authsvc-login-cookie"
	redirectParam = "redirect_uri"
	loginLifetime = 60 * 60 * 2 // 2 hours
)

// Options provides configuration options to the AuthenticationMiddleware.
type Options struct {
	Realm          string
	PublicRoots    []string
	LoginPath      string
	RequestChecker common.RequestChecker
}

// NewMiddleware returns a middlware suitable for authentication.
func NewMiddleware(root string, config *Options) common.Middleware {
	if config == nil {
		config = &Options{}
	}

	return &middleware{
		config: config,
	}
}

type middleware struct {
	config *Options
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
	if m.whitelisted(r) {
		n(w, r)
		return
	}
	if username := m.config.RequestChecker.IsAuthenticated(r); username != "" {
		n(w, r.WithContext(common.SetUsername(r.Context(), username)))
		return
	}
	m.unauthorized(w, r.URL, nil)
}

func (m *middleware) whitelisted(r *http.Request) bool {
	path := r.URL.Path
	for _, root := range m.config.PublicRoots {
		if strings.HasPrefix(path, root) {
			return true
		}
	}
	return false
}

func (m *middleware) unauthorized(w http.ResponseWriter, originalURL fmt.Stringer, err error) {
	var b strings.Builder
	_, _ = b.WriteString("authsvc")

	code := http.StatusUnauthorized
	parts := []string{}
	if m.config.Realm != "" {
		parts = append(parts, fmt.Sprintf("realm=%q", m.config.Realm))
	}
	switch err {
	case nil:
	default:
		errmsg := err.Error()
		parts = append(parts, fmt.Sprintf("error=%q", errmsg))
	}

	if len(parts) > 0 {
		_, _ = b.WriteString(" " + strings.Join(parts, ", "))
	}

	login := makeRedirect(m.config.LoginPath, originalURL)
	w.WriteHeader(code)
	w.Header().Set("WWW-Authenticate", b.String())
	w.Header().Set("Location", login.String())

	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="0;URL='%s'" /></head><body></body></html>`, login.String())
}

// LoginHandler returns a router that handles the login and logout routes.
func LoginHandler(authroot string, checker common.PasswordChecker, provider common.KeyProvider, insecure bool) http.Handler {
	sc := securecookie.New(provider.Hash(), provider.Block())
	h := &loginHandler{root: authroot, checker: checker, cookie: sc, insecure: insecure}

	r := mux.NewRouter()
	r.HandleFunc(authroot+loginPath, h.loginPOST).Methods("POST")
	r.HandleFunc(authroot+logoutPath, h.logoutPOST).Methods("POST")
	return r
}

type loginHandler struct {
	root     string
	checker  common.PasswordChecker
	cookie   *securecookie.SecureCookie
	insecure bool
}

func (m *loginHandler) loginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch button := r.Form.Get("submit"); button {
	case "Login":
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if !m.checker.IsAuthenticated(username, password) {
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

func (m *loginHandler) logoutPOST(w http.ResponseWriter, r *http.Request) {
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

func (m *loginHandler) setLoginCookie(username string, w http.ResponseWriter) {
	data := map[string]string{"username": username}
	switch v, err := m.cookie.Encode(cookieName, data); err {
	case nil:
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    v,
			Path:     "/",
			Expires:  time.Now().Add(time.Duration(loginLifetime) * time.Second),
			MaxAge:   loginLifetime,
			Secure:   !m.insecure,
			HttpOnly: true,
		})
	default:
		log.Printf("error encoding the cookie %q with username %q: %v", cookieName, username, err)
		common.JSONStatusResponse(http.StatusInternalServerError, w, "error logging in")
	}
}
func (m *loginHandler) clearLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
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
