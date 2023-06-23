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
	hkdf_secret         string
)

type Serializer interface {
	// returns master key
	// `[]byte` slice with length = 16
	// either truncate, or pad.
	MasterKey() []byte

	// Returns the masking []byte
	Mask() []byte

	// Write to struct
	// data: ciphertext or plaintext
	Serialize(data, nonce []byte)

	// Returns cipherText, nonce, salt
	// opposite of Serialize
	Deserialize() ([]byte, []byte)
}

func init() {
	hkdf_secret = os.Getenv("HKDF_SECRET")
	if hkdf_secret == "" {
		panic("`HKDF_SECRET` not set")
	}
}
