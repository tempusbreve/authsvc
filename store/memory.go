package store

import "time"

// NewMemoryCache implementes Cache with an in-memory store
func NewMemoryCache() Cache {
	return &memory{data: map[string]*val{}}
}

type clockFn func() time.Time

type val struct {
	key    string
	expire time.Time
	value  interface{}
}

type memory struct {
	now  clockFn
	data map[string]*val
}

func (m *memory) Put(key string, value interface{}) error {
	if m == nil || m.data == nil {
		return ErrInternal
	}
	m.data[key] = &val{key: key, value: value}
	return nil
}

func (m *memory) PutUntil(expire time.Time, key string, value interface{}) error {
	if m == nil || m.data == nil {
		return ErrInternal
	}
	m.data[key] = &val{key: key, expire: expire, value: value}
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
	if !v.expire.IsZero() && v.expire.Before(m.now()) {
		delete(m.data, key)
		return v.value, ErrExpired
	}
	return v.value, nil
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
