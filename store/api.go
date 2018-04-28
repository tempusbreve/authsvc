package store // import "breve.us/authsvc/store"

import (
	"errors"
	"time"
)

// Errors
var (
	ErrCacheMiss = errors.New("cache miss")
	ErrNotFound  = errors.New("not found")
	ErrExpired   = errors.New("expired")
	ErrInternal  = errors.New("internal error")
)

// Cache describes very simple TTL cache interface.
type Cache interface {
	Put(key string, value interface{}) error
	PutUntil(time time.Time, key string, value interface{}) error
	Get(key string) (interface{}, error)
	Delete(key string) error
}

type cacheValue struct {
	Key    string
	Expire time.Time
	Value  interface{}
}
