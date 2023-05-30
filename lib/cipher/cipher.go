package cipher

import (
	"crypto/aes"
	c "crypto/cipher"
	"os"
)

var block c.Block
var err error = nil

func init() {
	key := os.Getenv("CIPHER_KEY")
	block, err = aes.NewCipher([]byte(key))
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

func Decrypt(hashed string) string {
	buf := make([]byte, len(hashed))
	block.Decrypt(buf, []byte(hashed))
	return string(buf)
}
