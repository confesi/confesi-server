package cipher

import (
	"errors"
	"hash"
	"os"
)

const (
	MasterKeyLen = 16
)

var (
	ErrInvalidMasterKey = errors.New("invalid master key length")
	ErrInvalidKey       = errors.New("illegal key")
	ErrInvalidSalt      = errors.New("invalid salt")
	hkdf_secret         string
)

type Serializer interface {
	// length of the key has to be 32 bytes
	Key() []byte
}

type CipherResult struct {
	Ciphertext []byte
	Nonce      []byte
}

type KDF struct {
	algo func() hash.Hash
	salt []byte
}

func init() {
	hkdf_secret = os.Getenv("HKDF_SECRET")
	if hkdf_secret == "" {
		panic("`HKDF_SECRET` not set")
	}
}
