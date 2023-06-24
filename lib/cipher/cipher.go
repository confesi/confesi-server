package cipher

import (
	"errors"
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
	// returns master key
	// `[]byte` slice with length >= 16
	// either truncate, or pad.
	// use something unique, but not the masking data itself
	// ie: for masking user id, `MasterKey` can return `[]byte(userEmail)`
	MasterKey() []byte

	// Returns the masking data
	// ie `[]byte(user.id)`
	Mask() []byte

	// Write to struct
	// data: ciphertext or plaintext
	Serialize(data, nonce []byte)

	// Returns ciphertext, nonce
	// opposite of Serialize
	Deserialize() ([]byte, []byte)
}

func init() {
	hkdf_secret = os.Getenv("HKDF_SECRET")
	if hkdf_secret == "" {
		panic("`HKDF_SECRET` not set")
	}
}
