package common // import "breve.us/authsvc/common"

import (
	"encoding/base64"

	"github.com/gorilla/securecookie"
)

const (
	// HashKeySize is optimally 64 bytes
	HashKeySize = 64
	// BlockKeySize is optimally 32 bytes
	BlockKeySize = 32
)

// KeyProvider describes functions needed for encrypted cookies.
type KeyProvider interface {
	Hash() []byte
	Block() []byte
}

// DefaultKeyProvider creates random default values for the provider.
// This is only good for one execution run, unless the generated keys
// are persisted.
func DefaultKeyProvider() (KeyProvider, error) {
	return NewKeyProvider(Generate(HashKeySize), Generate(BlockKeySize))
}

// NewKeyProvider creates a default provider from base64 encoded values.
func NewKeyProvider(hash string, block string) (KeyProvider, error) {
	var (
		hashb, blockb []byte
		err           error
	)
	if hashb, err = Decode(hash); err != nil {
		return nil, err
	}
	if blockb, err = Decode(block); err != nil {
		return nil, err
	}
	return &defaultProvider{hash: hashb, block: blockb}, nil
}

// Generate creates a random key of keysize length, and returns it in
// base64 encoded text.
func Generate(keysize int) string {
	return base64.StdEncoding.
		EncodeToString(securecookie.GenerateRandomKey(keysize))
}

// Decode extracts a byte slice from a base64 encoded string.
func Decode(hash string) ([]byte, error) {
	return base64.StdEncoding.
		DecodeString(hash)
}

type defaultProvider struct {
	hash  []byte
	block []byte
}

func (s *defaultProvider) Hash() []byte {
	if s == nil {
		return nil
	}
	if len(s.hash) == 0 {
		s.hash = securecookie.GenerateRandomKey(HashKeySize)
	}
	return s.hash
}

func (s *defaultProvider) Block() []byte {
	if s == nil {
		return nil
	}
	if len(s.block) == 0 {
		s.block = securecookie.GenerateRandomKey(BlockKeySize)
	}
	return s.block
}
