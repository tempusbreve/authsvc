package authorization // import "breve.us/authsvc/authorization"

import (
	"errors"
	"time"

	"breve.us/authsvc/store"
)

var (
	errExpectString = errors.New("tokenCache expects the value to be a string")
	errExpectSlice  = errors.New("tokenCache expects the value list to be a []string")
)

func newTokenCache(clients store.Cache, tokens store.Cache) *tokenCache {
	return &tokenCache{clienttokens: clients, tokenclients: tokens}
}

type tokenCache struct {
	clienttokens store.Cache
	tokenclients store.Cache
}

// Put expects value to be a string token.
func (c *tokenCache) Put(key string, value interface{}) error {
	token, ok := value.(string)
	if !ok {
		return errExpectString
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
func (c *tokenCache) PutUntil(time time.Time, key string, value interface{}) error {
	token, ok := value.(string)
	if !ok {
		return errExpectString
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
func (c *tokenCache) Get(token string) (interface{}, error) {
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
		return v, errExpectString
	}
	return id, nil
}

// Delete removes token from the cache.
func (c *tokenCache) Delete(token string) error {
	id, err := c.getIDByToken(token)
	if err != nil {
		return err
	}
	return c.removeToken(id, token)
}

// getIDByToken returns client ID that correlates to the token.
func (c *tokenCache) getIDByToken(tok string) (string, error) {
	v, err := c.tokenclients.Get(tok)
	if err != nil {
		return "", err
	}
	if id, ok := v.(string); ok {
		return id, nil
	}
	return "", errExpectString
}

// GetTokenByID returns the token associated with a client ID.
func (c *tokenCache) getTokensByID(id string) ([]string, error) {
	v, err := c.Get(id)
	if err != nil {
		return nil, err
	}
	if list, ok := v.([]string); ok {
		return list, nil
	}
	return nil, errExpectSlice
}

func (c *tokenCache) addToken(id string, token string) error {
	toklist, err := c.getTokensByID(id)
	switch err {
	case nil, store.ErrNotFound:
	default:
		return err
	}
	toklist = append(toklist, token)
	return c.clienttokens.Put(id, toklist)
}

func (c *tokenCache) removeToken(id string, token string) error {
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
