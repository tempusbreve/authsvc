package store

import (
	"bytes"
	"encoding/gob"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// NewBoltDBCache implementes Cache with a BoltDB back-end;
func NewBoltDBCache(path string, bucket string) (Cache, error) {
	if bucket == "" {
		bucket = defaultBucket
	}
	cache := &bcache{
		path:   path,
		bucket: bucket,
	}
	switch db, err := cache.open(); err {
	case nil:
		err = db.Close()
		return cache, err
	default:
		return nil, err
	}
}

const (
	defaultBucket = "cache"
)

func init() {
	gob.Register(time.Now())
}

type bcache struct {
	sync.Once
	now    clockFn
	path   string
	bucket string
}

func (m *bcache) Put(key string, value interface{}) error {
	if m == nil {
		return ErrInternal
	}
	return m.put(key, &cacheValue{Key: key, Value: value})
}

func (m *bcache) PutUntil(expire time.Time, key string, value interface{}) error {
	if m == nil {
		return ErrInternal
	}
	return m.put(key, &cacheValue{Key: key, Expire: expire, Value: value})
}

func (m *bcache) Get(key string) (interface{}, error) {
	if m == nil {
		return nil, ErrInternal
	}
	v, err := m.get(key)
	if err != nil {
		return nil, err
	}
	if m.now == nil {
		m.now = time.Now
	}
	if !v.Expire.IsZero() && v.Expire.Before(m.now()) {
		_ = m.remove(key)
		return v.Value, ErrExpired
	}
	return v.Value, nil
}

func (m *bcache) Delete(key string) error {
	if m == nil {
		return ErrInternal
	}
	return m.remove(key)
}

func (m *bcache) Keys() []string {
	var keys []string

	db, err := m.open()
	if err != nil {
		return keys
	}
	defer func() { _ = db.Close() }()

	if err := db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(m.bucket)).ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	}); err != nil {
		log.Printf("error iterating keys: %v", err)
	}
	sort.Strings(keys)
	return keys
}

func (m *bcache) put(key string, v *cacheValue) error {
	db, err := m.open()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(v); err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error { return tx.Bucket([]byte(m.bucket)).Put([]byte(key), buf.Bytes()) })
}

func (m *bcache) get(key string) (*cacheValue, error) {
	db, err := m.open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()
	v := &cacheValue{}
	if err = db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte(m.bucket)).Get([]byte(key))
		if data == nil {
			return ErrNotFound
		}
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		return dec.Decode(v)
	}); err != nil {
		return nil, err
	}
	return v, nil
}

func (m *bcache) remove(key string) error {
	db, err := m.open()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return db.Update(deleteKey(m.bucket, key))
}

func (m *bcache) open() (*bolt.DB, error) {
	db, err := bolt.Open(m.path, 0600, nil)
	if err == nil {
		m.Do(func() {
			err = db.Update(createBucket(m.bucket))
		})
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createBucket(bucket string) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}
}

func deleteKey(bucket string, key string) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error { return tx.Bucket([]byte(bucket)).Delete([]byte(key)) }
}
