package client // import "breve.us/authsvc/client"

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"time"

	"breve.us/authsvc/store"
)

// Errors
var (
	ErrNotFound = errors.New("client id not found")
)

func init() {
	gob.Register(&Details{})
}

// Details represents a registered client application
type Details struct {
	// ID is the unique identifier for the client
	ID string `json:"id"`
	// Name is the user friendly name of the client application
	Name string `json:"name"`
	// Endpoints are the list of approved callback endpoints
	Endpoints []string `json:"endpoints"`
}

// Registry is the manager for all registered clients
type Registry struct {
	cache store.Cache
}

// NewRegistry returns an initialized ClientRegistry
func NewRegistry(cache store.Cache) *Registry { return &Registry{cache: cache} }

// VerifyClient returns true if the client is a registered client
func (c *Registry) VerifyClient(client string) bool {
	_, err := c.Get(client)
	return err == nil
}

// VerifyRedirect returns true if the client is a registered client, and
// the redirect uri is an approved uri
func (c *Registry) VerifyRedirect(client string, redirect string) bool {
	cl, err := c.Get(client)
	if err != nil {
		return false
	}
	for _, redir := range cl.Endpoints {
		if redirect == redir {
			return true
		}
	}
	return false
}

// Get returns a client registration by id, or an error if not found
func (c *Registry) Get(id string) (*Details, error) {
	v, err := c.cache.Get(id)
	if err != nil {
		return nil, err
	}
	if client, ok := v.(*Details); ok {
		return client, nil
	}
	return nil, ErrNotFound
}

// Put save the client registration
func (c *Registry) Put(client *Details) error { return c.cache.Put(client.ID, client) }

// PutUntil saves the client registration until the expiration date
func (c *Registry) PutUntil(expire time.Time, client *Details) error {
	return c.cache.PutUntil(expire, client.ID, client)
}

// Delete removes the client registration
func (c *Registry) Delete(id string) error {
	return c.cache.Delete(id)
}

// LoadFromJSON loads clients encoded in JSON
func (c *Registry) LoadFromJSON(r io.Reader) error {
	var (
		clientSlice []Details
		client      Details
	)
	dec := json.NewDecoder(r)
	if dec.Decode(&clientSlice) != nil {
		if err := dec.Decode(&client); err != nil {
			return err
		}
		clientSlice = append(clientSlice, client)
	}
	for _, cl := range clientSlice {
		if err := c.Put(&cl); err != nil {
			return err
		}
	}
	return nil
}

// SaveToJSON stores the registry clients to the io.Writer as JSON
func (c *Registry) SaveToJSON(w io.Writer) error {
	var clients []Details
	keys, err := c.cache.Keys()
	if err != nil {
		return err
	}
	for _, id := range keys {
		if v, err := c.cache.Get(id); err == nil {
			if client, ok := v.(*Details); ok {
				clients = append(clients, *client)
			}
		} else {
			return err
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(clients)
}
