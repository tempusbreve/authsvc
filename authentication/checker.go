package authentication // import "breve.us/authsvc/authentication"

import (
	"log"
	"net/http"

	"github.com/gorilla/securecookie"

	"breve.us/authsvc/common"
	"breve.us/authsvc/user"
)

// NewSecureCookieChecker verifies authentication tokens on requests according to secure cookies
func NewSecureCookieChecker(provider common.KeyProvider, users *user.Registry) common.RequestChecker {
	sc := securecookie.New(provider.Hash(), provider.Block())
	return &cookieChecker{sc: sc, users: users}
}

type cookieChecker struct {
	sc    *securecookie.SecureCookie
	users *user.Registry
}

func (cc *cookieChecker) IsAuthenticated(r *http.Request) string {
	log.Printf("CookieChecker.IsAuthenticated()...")
	if c, err := r.Cookie(cookieName); err == nil {
		var data map[string]string
		if err = cc.sc.Decode(cookieName, c.Value, &data); err == nil {
			if username, ok := data["username"]; ok {
				if d, err := cc.users.Get(username); err == nil {
					log.Printf("CookieChecker.IsAuthenticated() details: %+v", d)
					if d.State == "active" {
						return username
					}
				}
			}
		}
	}
	return ""
}
