package oauth // import "breve.us/authsvc/oauth"

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"time"

	"breve.us/authsvc/store"
)

func init() {
	gob.Register(&Client{})
}

// Client represents a registered client application
type Client struct {
	// ID is the unique identifier for the client
	ID string `json:"id"`
	// Name is the user friendly name of the client application
	Name string `json:"name"`
	// Endpoints are the list of approved callback endpoints
	Endpoints []string `json:"endpoints"`
}

// ClientRegistry is the manager for all registered clients
type ClientRegistry struct {
	cache store.Cache
}

// NewClientRegistry returns an initialized ClientRegistry
func NewClientRegistry(cache store.Cache) *ClientRegistry { return &ClientRegistry{cache: cache} }

// VerifyClient returns true if the client is a registered client
func (c *ClientRegistry) VerifyClient(client string) bool {
	_, err := c.Get(client)
	return err == nil
}

// VerifyRedirect returns true if the client is a registered client, and
// the redirect uri is an approved uri
func (c *ClientRegistry) VerifyRedirect(client string, redirect string) bool {
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
func (c *ClientRegistry) Get(id string) (*Client, error) {
	v, err := c.cache.Get(id)
	if err != nil {
		return nil, err
	}
	if client, ok := v.(*Client); ok {
		return client, nil
	}
	return nil, ErrInvalidClient
}

// Put save the client registration
func (c *ClientRegistry) Put(client *Client) error { return c.cache.Put(client.ID, client) }

// PutUntil saves the client registration until the expiration date
func (c *ClientRegistry) PutUntil(expire time.Time, client *Client) error {
	return c.cache.PutUntil(expire, client.ID, client)
}

// Delete removes the client registration
func (c *ClientRegistry) Delete(id string) error {
	return c.cache.Delete(id)
}

// LoadFromJSON loads clients encoded in JSON
func (c *ClientRegistry) LoadFromJSON(r io.Reader) error {
	var (
		clientSlice []Client
		client      Client
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
func (c *ClientRegistry) SaveToJSON(w io.Writer) error {
	var clients []Client
	for _, id := range c.cache.Keys() {
		if v, err := c.cache.Get(id); err == nil {
			if client, ok := v.(*Client); ok {
				clients = append(clients, *client)
			}
		} else {
			return err
		}
	}
	enc := json.NewEncoder(w)
	return enc.Encode(clients)
}
