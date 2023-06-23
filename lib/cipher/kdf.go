package cipher

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"
)

const (
	MasterKeyLen = 16
)

var (
	ErrInvalidMasterKey = errors.New("invalid master key length")
	hkdf_secret         string
)

// Key deviration function
type KDF struct {
	algo func() hash.Hash
	salt []byte
}

type Serializer interface {
	// returns master key
	// `[]byte` slice with length = 16
	// either truncate, or pad.
	MasterKey() []byte

	// Write to struct the key derived from master key
	Mask(key []byte)
}

func init() {
	hkdf_secret = os.Getenv("HKDF_SECRET")
	if hkdf_secret == "" {
		panic("`HKDF_SECRET` not set")
	}
}

func NewKDF() (*KDF, error) {
	algo := sha256.New
	salt := make([]byte, algo().Size())
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return &KDF{algo, salt}, nil
}

func NewWithSalt(saltStr string) (*KDF, error) {
	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return nil, err
	}
	return &KDF{sha256.New, salt}, nil
}

func (kdf *KDF) Salt() string {
	return base64.StdEncoding.EncodeToString(kdf.salt)
}

func (kdf *KDF) GenKey(data Serializer) ([]byte, error) {
	p := data.MasterKey()
	if len(p) != 16 {
		return nil, ErrInvalidMasterKey
	}

	secret := []byte(hkdf_secret)
	p = append(p, secret...)

	result := hkdf.New(kdf.algo, []byte(p), kdf.salt, nil)
	key := make([]byte, len(p))
	if _, err := result.Read(key); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := hex.NewEncoder(&buf)
	if _, err := w.Write(key); err != nil {
		return nil, err
	}

	return buf.Bytes()[:32], nil
}
