package authorization // import "breve.us/authsvc/authorization"

import (
	"errors"
	"time"

	"breve.us/authsvc/store"
)

// Errors
var (
	ErrExpectStringValue = errors.New("TokenCache expects the value to be a string")
	ErrExpectStringSlice = errors.New("TokenCache expects the value list to be a []string")
)

// NewTokenCache creates a new TokenCache from the provided store.Cache's
// representing client tokens, and token clients.
func NewTokenCache(clients store.Cache, tokens store.Cache) *TokenCache {
	return &TokenCache{clienttokens: clients, tokenclients: tokens}
}

// TokenCache provides a forward/backward lookup cache for clients and
// tokens.
//
// There is an n->1 relationship between tokens and client id, so this
// cache abstracts managing the lookup and expiration of tokens per
// client id.
type TokenCache struct {
	clienttokens store.Cache
	tokenclients store.Cache
}

// Put expects value to be a string token.
func (c *TokenCache) Put(key string, value interface{}) error {
	token, ok := value.(string)
	if !ok {
		return ErrExpectStringValue
	}
	if err := c.addToken(key, token); err != nil {
		return err
	}
	if err := c.tokenclients.Put(token, key); err != nil {
		return err
	}
	return nil
}

// PutUntil expects value to be a string token.
func (c *TokenCache) PutUntil(time time.Time, key string, value interface{}) error {
	token, ok := value.(string)
	if !ok {
		return ErrExpectStringValue
	}
	if err := c.addToken(key, token); err != nil {
		return err
	}
	if err := c.tokenclients.PutUntil(time, token, key); err != nil {
		return err
	}
	return nil
}

// Get expects to be given a token, and to return a client id.
func (c *TokenCache) Get(token string) (interface{}, error) {
	v, err := c.tokenclients.Get(token)
	id, ok := v.(string)
	switch err {
	case nil:
	case store.ErrExpired:
		if ok {
			_ = c.removeToken(id, token)
		}
		fallthrough
	default:
		return v, err
	}
	if !ok {
		return v, ErrExpectStringValue
	}
	return id, nil
}

// Delete removes token from the cache.
func (c *TokenCache) Delete(token string) error {
	id, err := c.getIDByToken(token)
	if err != nil {
		return err
	}
	return c.removeToken(id, token)
}

// getIDByToken returns client ID that correlates to the token.
func (c *TokenCache) getIDByToken(tok string) (string, error) {
	v, err := c.tokenclients.Get(tok)
	if err != nil {
		return "", err
	}
	if id, ok := v.(string); ok {
		return id, nil
	}
	return "", ErrExpectStringValue
}

// GetTokenByID returns the token associated with a client ID.
func (c *TokenCache) getTokensByID(id string) ([]string, error) {
	v, err := c.Get(id)
	if err != nil {
		return nil, err
	}
	if list, ok := v.([]string); ok {
		return list, nil
	}
	return nil, ErrExpectStringSlice
}

func (c *TokenCache) addToken(id string, token string) error {
	toklist, err := c.getTokensByID(id)
	switch err {
	case nil, store.ErrNotFound:
	default:
		return err
	}
	toklist = append(toklist, token)
	return c.clienttokens.Put(id, toklist)
}

func (c *TokenCache) removeToken(id string, token string) error {
	_ = c.tokenclients.Delete(token)
	toklist, err := c.getTokensByID(id)
	if err != nil {
		return nil
	}
	var newlist []string
	for _, tok := range toklist {
		if tok != token {
			newlist = append(newlist, tok)
		}
	}
	return c.clienttokens.Put(id, newlist)
}
