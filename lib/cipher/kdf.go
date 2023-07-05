package cipher

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

func NewKDF() (*KDF, error) {
	algo := sha256.New
	salt := make([]byte, algo().Size())
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return &KDF{algo, salt}, nil
}

func (kdf *KDF) GenKey(data Serializer) ([]byte, error) {
	masterKey := data.Key()
	if len(masterKey) < MasterKeyLen {
		return nil, ErrInvalidMasterKey
	}
	if kdf.salt == nil {
		return nil, ErrInvalidSalt
	}

	masterKey = append(masterKey, []byte(hkdf_secret)...)
	result := hkdf.New(kdf.algo, masterKey, kdf.salt, nil)

	key := make([]byte, 32)
	if _, err := result.Read(key); err != nil {
		return nil, err
	}

	return key, nil

}
