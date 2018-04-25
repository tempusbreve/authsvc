package authsvc // import "breve.us/authsvc"

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"github.com/gorilla/securecookie"
)

type Seeder interface {
	HashKey() []byte
	BlockKey() []byte
}

func NewSeeder() Seeder { return &defseeder{} }

func NewFileSeeder(name string) (Seeder, error) {
	var err error
	var data []byte
	if data, err = ioutil.ReadFile(name); err != nil {
		return nil, err
	}
	seeder := &fileseeder{}
	if err = json.Unmarshal(data, seeder); err != nil {
		return nil, err
	}
	return seeder, nil
}

const (
	hashKeySize  = 64
	blockKeySize = 32
)

type fileseeder struct {
	Hash  string `json:"hash"`
	Block string `json:"block"`
}

func (f *fileseeder) HashKey() []byte {
	if data, err := base64.StdEncoding.DecodeString(f.Hash); err == nil {
		return data
	}
	hash := securecookie.GenerateRandomKey(hashKeySize)
	f.Hash = base64.StdEncoding.EncodeToString(hash)
	return hash
}

func (f *fileseeder) BlockKey() []byte {
	if data, err := base64.StdEncoding.DecodeString(f.Block); err == nil {
		return data
	}
	block := securecookie.GenerateRandomKey(blockKeySize)
	f.Block = base64.StdEncoding.EncodeToString(block)
	return block
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
