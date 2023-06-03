package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"os"
)

func Encrypt(plainText string) string {
	aesgcm, nonce := getAesGCM()
	encrypted := aesgcm.Seal(nil, nonce, []byte(plainText), nil)
	return string(encrypted)
}

func Decrypt(encryptedText string) (string, error) {
	aesgcm, nonce := getAesGCM()
	plainText, err := aesgcm.Open(nil, nonce, []byte(encryptedText), nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// NOTE: `cipher.Block` is not threadsafe, initialize a new one every
// function call instead of a global `init()`
func getAesGCM() (cipher.AEAD, []byte) {
	key := os.Getenv("CIPHER_KEY")
	if key == "" {
		panic("`CIPHER_KEY` env not set")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		// NOTE: valid key has the length of 16, 24, or 32 bytes to generate
		// a AES-128, AES-192 or AES-256 respectively.
		// and `panic(err)` since err is returned when they key has an invalid
		// length (ie, not of 16, 24, or 32 bytes.)
		panic(err)
	}

	nonceHex := os.Getenv("CIPHER_NONCE")
	if nonceHex == "" {
		panic("`CIPHER_NONCE` env not set")
	}

	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		panic("invalid `CIPHER_NONCE`:" + nonceHex)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	return aesgcm, nonce
}
