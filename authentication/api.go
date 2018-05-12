package authentication // import "breve.us/authsvc/authentication"

import (
	"golang.org/x/crypto/bcrypt"
)

// Checker describes functionality to verify passwords
type Checker interface {
	Authenticate(username string, password string) bool
}

// NewBasicChecker returns a default checker based on a map of username/bcrypt-hashed-password
func NewBasicChecker(creds map[string]string) Checker {
	if creds == nil {
		creds = map[string]string{"jweldon": "$2a$10$hybo9Mikt1PfpSDUiivghuuSKQYamE8Or5ADCuA1Su5u7BLnThvpK"}
	}
	return &def{creds: creds}
}

type def struct {
	creds map[string]string
}

func (d *def) Authenticate(username string, password string) bool {
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
