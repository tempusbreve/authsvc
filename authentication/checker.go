package authentication // import "breve.us/authsvc/authentication"

import (
	"log"
	"net/http"

	"breve.us/authsvc/common"
	"breve.us/authsvc/user"
	"github.com/gorilla/securecookie"
)

func cookieRequestChecker(sc *securecookie.SecureCookie, users *user.Registry) common.RequestChecker {
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
