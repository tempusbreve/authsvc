package user // import "breve.us/authsvc/user"

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"

	"breve.us/authsvc/store"
)

// Common Errors
var (
	ErrInvalidUser = errors.New("invalid user")
)

func init() {
	gob.Register(&Details{})
}

// Details describes the user details
type Details struct {
	ID       int
	Username string
	Password string
	Email    string
	Name     string
	State    string
}

// Registry maintains the known users
type Registry struct {
	cache store.Cache
}

// NewUserRegistry returns an initialized UserRegistry
func NewUserRegistry(cache store.Cache) *Registry { return &Registry{cache: cache} }

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
	for _, id := range u.cache.Keys() {
		if v, err := u.cache.Get(id); err == nil {
			if user, ok := v.(*Details); ok {
				users = append(users, *user)
			}
		} else {
			return err
		}
	}
	return json.NewEncoder(w).Encode(users)
}
