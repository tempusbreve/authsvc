package store

import (
	"sort"
	"time"
)

// NewMemoryCache implementes Cache with an in-memory store
func NewMemoryCache() Cache {
	return &memory{data: map[string]*cacheValue{}}
}

type clockFn func() time.Time

type memory struct {
	now  clockFn
	data map[string]*cacheValue
}

func (m *memory) Put(key string, value interface{}) error {
	if m == nil || m.data == nil {
		return ErrInternal
	}
	m.data[key] = &cacheValue{Key: key, Value: value}
	return nil
}

func (m *memory) PutUntil(expire time.Time, key string, value interface{}) error {
	if m == nil || m.data == nil {
		return ErrInternal
	}
	m.data[key] = &cacheValue{Key: key, Expire: expire, Value: value}
	return nil
}

func (m *memory) Get(key string) (interface{}, error) {
	if m == nil || m.data == nil {
		return nil, ErrInternal
	}
	v, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	if m.now == nil {
		m.now = time.Now
	}
	if !v.Expire.IsZero() && v.Expire.Before(m.now()) {
		delete(m.data, key)
		return v.Value, ErrExpired
	}
	return v.Value, nil
}

func (m *memory) Delete(key string) error {
	if m == nil || m.data == nil {
		return ErrInternal
	}
	_, ok := m.data[key]
	if !ok {
		return ErrNotFound
	}
	delete(m.data, key)
	return nil
}

func (m *memory) Keys() []string {
	var keys []string
	if m != nil {
		for k := range m.data {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
