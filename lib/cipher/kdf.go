package cipher

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Key deviration function
type KDF struct {
	algo func() hash.Hash
	salt []byte
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
	masterKey := data.MasterKey()
	if len(masterKey) < MasterKeyLen {
		return nil, ErrInvalidMasterKey
	}
	if kdf.salt == nil {
		return nil, ErrInvalidSalt
	}

	masterKey = append(masterKey, []byte(hkdf_secret)...)

	result := hkdf.New(kdf.algo, []byte(masterKey), kdf.salt, nil)
	key := make([]byte, len(masterKey))
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
