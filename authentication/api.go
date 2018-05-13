package authentication // import "breve.us/authsvc/authentication"

import (
	"log"

	"golang.org/x/crypto/bcrypt"

	"breve.us/authsvc/common"
)

// NewBasicChecker returns a default checker based on a map of username/bcrypt-hashed-password
func NewBasicChecker(creds map[string]string) common.PasswordChecker {
	if creds == nil {
		creds = map[string]string{"jweldon": "$2a$10$hybo9Mikt1PfpSDUiivghuuSKQYamE8Or5ADCuA1Su5u7BLnThvpK"}
	}
	return &def{creds: creds}
}

type def struct {
	creds map[string]string
}

func (d *def) IsAuthenticated(username string, password string) bool {
	log.Printf("BasicChecker.IsAuthenticated(%q,%q)", username, password)

	if username == "" || password == "" {
		return false
	}
	p, ok := d.creds[username]
	if !ok {
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(p), []byte(password)) == nil {
		return true
	}
	return false
}
