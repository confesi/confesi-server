package cipher

import (
	"crypto/aes"
	c "crypto/cipher"
	"encoding/hex"
)

var block c.Block

var err error = nil

func init() {
	// key := os.Getenv("CIPHER_KEY")
	key := "thisis32bitlongpassphraseimusing"
	block, _ = aes.NewCipher([]byte(key))
	if err != nil {
		// panic since err here signifies an invalid key.
		// key can only be of length 16, 24, or 32 bytes.
		// https://pkg.go.dev/crypto/aes#NewCipher
		panic(err)
	}
}

func Encrypt(plainText string) string {
	buf := make([]byte, len(plainText))
	block.Encrypt(buf, []byte(plainText))
	return string(buf)
}

func Decrypt(hashed string) (string, error) {
	cipherText, err := hex.DecodeString(hashed)
	if err != nil {
		return "", err
	}

	buf := make([]byte, len(cipherText))
	block.Decrypt(buf, cipherText)
	return string(buf), nil
}
