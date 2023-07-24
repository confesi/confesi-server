package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

// a 32 byte secret for ciphering, will panic if it has a none 32 length
var key []byte

func init() {
	k := os.Getenv("CIPHER_KEY")
	if k == "" {
		panic("`CIPHER_KEY` not set")
	}

	key = []byte(k)
	if len(key) != 32 {
		panic("invalid key length")
	}
}

// `ad` must be the same for both Cipher and Decipher.it i
// stands for Additional Data, use something unique to it (ie: user id)
func Cipher(plaintext []byte, ad []byte) ([]byte, error) {
	if len(ad) == 0 {
		return nil, fmt.Errorf("invalid length for additional data: %d", len(ad))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	c, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, c.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	cipherSize := len(plaintext) + c.NonceSize() + c.Overhead()
	ciphertext := make([]byte, 0, cipherSize)

	ciphertext = append(ciphertext, nonce...)
	ciphertext = c.Seal(ciphertext, nonce, plaintext, ad)

	return ciphertext, nil
}

// `ad` must be the same for both Cipher and Decipher.it i
// stands for Additional Data, use something unique to it (ie: user id)
func Decipher(ciphertext []byte, ad []byte) ([]byte, error) {
	if len(ad) == 0 {
		return nil, fmt.Errorf("invalid length for additional data: %d", len(ad))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	c, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := c.NonceSize()
	ptSize := len(ciphertext) - nonceSize - c.Overhead()
	pt := make([]byte, 0, ptSize)

	nonce := ciphertext[:nonceSize]
	cipherdata := ciphertext[nonceSize:]
	pt, err = c.Open(pt, nonce, cipherdata, ad)
	if err != nil {
		return nil, err
	}

	return pt, nil
}
