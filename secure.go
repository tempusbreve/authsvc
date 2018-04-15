package authsvc // import "breve.us/authsvc"

import (
	"github.com/gorilla/securecookie"
)

type Seeder interface {
	HashKey() []byte
	BlockKey() []byte
}

const (
	hashKeySize  = 64
	blockKeySize = 32
)

type defseeder struct {
	hash  []byte
	block []byte
}

func (s *defseeder) HashKey() []byte {
	if s == nil {
		return nil
	}
	if len(s.hash) == 0 {
		s.hash = securecookie.GenerateRandomKey(hashKeySize)
	}
	return s.hash
}

func (s *defseeder) BlockKey() []byte {
	if s == nil {
		return nil
	}
	if len(s.block) == 0 {
		s.block = securecookie.GenerateRandomKey(blockKeySize)
	}
	return s.block
}
