package authsvc // import "breve.us/authsvc"

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alecthomas/template"
	"github.com/gorilla/securecookie"
)

func newLoginHandler(seeder Seeder) http.Handler {
	if seeder == nil {
		seeder = &defseeder{}
	}
	return &login{
		gen: securecookie.New(seeder.HashKey(), seeder.BlockKey()),
	}
}

type login struct {
	gen *securecookie.SecureCookie
}

func (l *login) serveForm(w http.ResponseWriter, r *http.Request) {
	form := loginForm
	if l.loggedIn(r) {
		form = logoutForm
	}
	if err := form.Execute(w, nil); err != nil {
		writeJSONCode(http.StatusInternalServerError, w, nil)
	}
}

func (l *login) clearLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func (l *login) setLoginCookie(username string, w http.ResponseWriter) {
	data := map[string]string{"username": username}
	switch v, err := l.gen.Encode(cookieName, data); err {
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

func (l *login) loggedIn(r *http.Request) bool {
	switch c, err := r.Cookie(cookieName); err {
	case nil:
		data := map[string]string{}
		if err = l.gen.Decode(cookieName, c.Value, &data); err == nil {
			// TODO: verify username (locked out, etc)
			return true
		}
	default:
	}
	return false
}

func (l *login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		l.serveForm(w, r)
	case http.MethodPost:
		l.handleLogin(w, r)
	default:
		writeJSONCode(http.StatusBadRequest, w, fmt.Sprintf("method %q not supported", r.Method))
	}
}

func (l *login) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONCode(http.StatusBadRequest, w, err.Error())
		return
	}
	switch button := r.Form.Get("submit"); button {
	case "Logout":
		l.clearLoginCookie(w)
		writeRedirect(w, r, "/", map[string]string{"msg": "logged out"})
	case "Login":
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if !l.validate(username, password) {
			writeJSONCode(http.StatusBadRequest, w, "bad username or password")
			return
		}
		l.setLoginCookie(username, w)
		writeRedirect(w, r, "/", map[string]string{"msg": "logged in"})
	default:
		writeJSONCode(http.StatusBadRequest, w, fmt.Sprintf("bad button %q", button))
	}
}

func (l *login) validate(username, password string) bool {
	if username == "" || password == "" {
		return false
	}
	if username == "jweldon" && password == "password" {
		return true
	}
	return false
}

var (
	cookieName    = "authsvc-login-cookie"
	loginLifetime = 60 * 60 * 2 // 2 hours
	loginForm     = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <title>Login</title>
  </head>
  <body>
    <h3>Login</h3>
    <form method="post">
      <div>
        <label for="username">Username: </label>
        <input type="text" name="username" id="username" placeholder="username" required>
        <label for="password">Password: </label>
        <input type="password" name="password" id="password" autocomplete="current-password" required>
        <input type="submit" name="submit" value="Login">
      </div>
    </form>
  </body>
</html>
`))
	logoutForm = template.Must(template.New("logout").Parse(`<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <title>Logout</title>
  </head>
  <body>
    <h3>Logout</h3>
    <form method="post">
      <div>
        <input type="submit" name="submit" value="Logout">
      </div>
    </form>
  </body>
</html>
`))
)
