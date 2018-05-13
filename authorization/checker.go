package authorization // import "breve.us/authsvc/authorization"

import (
	"log"
	"net/http"
	"strings"

	"breve.us/authsvc/common"
	"breve.us/authsvc/user"
)

// NewRequestChecker returns a Bearer Token request checker
func NewRequestChecker(cache *TokenCache, users *user.Registry) common.RequestChecker {
	return &requestChecker{tc: cache, ur: users}
}

type requestChecker struct {
	tc *TokenCache
	ur *user.Registry
}

func (c *requestChecker) IsAuthenticated(r *http.Request) string {
	log.Printf("Bearer RequestChecker.IsAuthenticated()...")
	if tok := bearerToken(r); tok != "" {
		log.Printf("Bearer RequestChecker.IsAuthenticated() Token: %q", tok)
		if u, err := c.tc.Get(tok); err == nil {
			if username, ok := u.(string); ok {
				log.Printf("Bearer RequestChecker.IsAuthenticated() Username: %q", username)
				if d, err := c.ur.Get(username); err == nil {
					log.Printf("Bearer RequestChecker.IsAuthenticated() details: %+v", d)
					if d.State == "active" {
						return username
					}
				} else {
					log.Printf("ERROR: Bearer RequestChecker - registry Get %v", err)
				}
			}
		} else {
			log.Printf("ERROR: Bearer RequestChecker - token cache Get %v", err)
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