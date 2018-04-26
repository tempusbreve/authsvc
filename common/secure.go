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

// Seeder describes functions needed for encrypted cookies.
type Seeder interface {
	HashKey() []byte
	BlockKey() []byte
}

// NewDefaultSeeder creates random default values for the Seeder. This
// is only good for one execution run, unless the generated keys are
// persisted.
func NewDefaultSeeder() (Seeder, error) {
	return NewSeeder(Generate(HashKeySize), Generate(BlockKeySize))
}

// NewSeeder creates a default Seeder from base64 encoded seed values.
func NewSeeder(hash string, block string) (Seeder, error) {
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
	return &defseeder{hash: hashb, block: blockb}, nil
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

type defseeder struct {
	hash  []byte
	block []byte
}

func (s *defseeder) HashKey() []byte {
	if s == nil {
		return nil
	}
	if len(s.hash) == 0 {
		s.hash = securecookie.GenerateRandomKey(HashKeySize)
	}
	return s.hash
}

func (s *defseeder) BlockKey() []byte {
	if s == nil {
		return nil
	}
	if len(s.block) == 0 {
		s.block = securecookie.GenerateRandomKey(BlockKeySize)
	}
	return s.block
}
