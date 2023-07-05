package cipher

import (
	"crypto/aes"
	c "crypto/cipher"
	"crypto/rand"
	"io"
)

func Cipher(plaintext []byte, key []byte) (*CipherResult, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesGcm, err := c.NewGCM(block)
	if err != nil {
		return nil, err
	}

	cipher := aesGcm.Seal(nil, nonce, plaintext, nil)

	return &CipherResult{cipher, nonce}, nil
}

func (ciphertext *CipherResult) Decipher(key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesGcm, err := c.NewGCM(block)
	if err != nil {
		return err
	}

	plainText, err := aesGcm.Open(nil, ciphertext.Nonce, ciphertext.Ciphertext, nil)
	if err != nil {
		return err
	}
	ciphertext.Ciphertext = plainText

	return nil
}
