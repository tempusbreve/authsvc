package user // import "breve.us/authsvc/user"

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"log"

	"golang.org/x/crypto/bcrypt"

	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
)

// Common Errors
var (
	ErrInvalidUser = errors.New("invalid user")
)

func init() {
	gob.Register(&Details{})
}

// State describes User States
type State string

// Valid User States
const (
	Active   State = "active"
	Inactive State = "inactive"
)

// Details describes the user details
type Details struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	State    State  `json:"state"`
}

func (d *Details) toFilteredMap() map[string]interface{} {
	return map[string]interface{}{
		"id":       d.ID,
		"username": d.Username,
		"login":    d.Username,
		"email":    d.Email,
		"name":     d.Name,
	}
}

// Registry maintains the known users
type Registry struct {
	cache store.Cache
}

// NewRegistry returns an initialized UserRegistry
func NewRegistry(cache store.Cache) *Registry { return &Registry{cache: cache} }

// Get returns a user registration by username, or an error if not found
func (u *Registry) Get(username string) (*Details, error) {
	v, err := u.cache.Get(username)
	if err != nil {
		return nil, err
	}
	if user, ok := v.(*Details); ok {
		return user, nil
	}
	return nil, ErrInvalidUser
}

// Put save the user registration
func (u *Registry) Put(user *Details) error { return u.cache.Put(user.Username, user) }

// Delete removes the user registration
func (u *Registry) Delete(username string) error { return u.cache.Delete(username) }

// LoadFromJSON loads users encoded in JSON
func (u *Registry) LoadFromJSON(r io.Reader) error {
	var (
		userSlice []Details
		user      Details
	)
	dec := json.NewDecoder(r)
	if dec.Decode(&userSlice) != nil {
		if err := dec.Decode(&user); err != nil {
			return err
		}
		userSlice = append(userSlice, user)
	}
	for _, cl := range userSlice {
		if err := u.Put(&cl); err != nil {
			return err
		}
	}
	return nil
}

// SaveToJSON stores the registry users to the io.Writer as JSON
func (u *Registry) SaveToJSON(w io.Writer) error {
	var users []Details
	keys, err := u.cache.Keys()
	if err != nil {
		return err
	}
	for _, id := range keys {
		if v, err := u.cache.Get(id); err == nil {
			if user, ok := v.(*Details); ok {
				users = append(users, *user)
			}
		} else {
			return err
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(users)
}

// BcryptChecker creates in implementation of common.Checker that uses
// bcrypt to verify password against the password field in the user
// details
func (u *Registry) BcryptChecker() common.PasswordChecker {

	return &bchecker{r: u}
}

type bchecker struct {
	r *Registry
}

// IsAuthenticated requires that neither username nor password is empty,
// the username is valid, the stored user state is Active, the
// stored password is not empty, and that the supplied password matches
// the stored password when compared with a bcrypt algorithm
func (b *bchecker) IsAuthenticated(username, password string) bool {
	log.Printf("Bcrypt PasswordChecker.IsAuthenticated(%q, %q)", username, password)
	if username == "" || password == "" {
		return false
	}
	u, err := b.r.Get(username)
	if err != nil {
		return false
	}
	if u.State != Active {
		return false
	}
	if u.Password == "" {
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil {
		return true
	}
	return false
}

// PlainTextChecker creates in implementation of common.Checker that uses
// simple text comparison to verify password against the plain text
// password field in the user details
func (u *Registry) PlainTextChecker() common.PasswordChecker {
	return &pchecker{r: u}
}

type pchecker struct {
	r *Registry
}

// IsAuthenticated requires that neither username nor password is empty,
// the username is valid, the stored user state is Active, the
// stored password is not empty, and that the supplied password matches
// the stored password when compared as plain text strings
func (p *pchecker) IsAuthenticated(username, password string) bool {
	log.Printf("Plain PasswordChecker.IsAuthenticated(%q, %q)", username, password)
	if username == "" || password == "" {
		return false
	}
	u, err := p.r.Get(username)
	if err != nil {
		return false
	}
	if u.State != Active {
		return false
	}
	if u.Password == "" {
		return false
	}
	if password == u.Password {
		return true
	}
	return false
}
