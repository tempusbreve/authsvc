package authorization // import "breve.us/authsvc/authorization"

import (
	"net/http"
	"strings"

	"breve.us/authsvc/common"
	"breve.us/authsvc/user"
)

func newTokenRequestChecker(cache *tokenCache, users *user.Registry) common.RequestChecker {
	return &requestChecker{tc: cache, ur: users}
}

type requestChecker struct {
	tc *tokenCache
	ur *user.Registry
}

func (c *requestChecker) IsAuthenticated(r *http.Request) string {
	if tok := bearerToken(r); tok != "" {
		if u, err := c.tc.Get(tok); err == nil {
			if username, ok := u.(string); ok {
				if d, err2 := c.ur.Get(username); err2 == nil {
					if d.State == "active" {
						return username
					}
				} else {
				}
			}
		}
	}
	return ""
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}
