package masking

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
)

var secretKey []byte

func init() {
	// load from .env
	m := os.Getenv("MASK_SECRET")
	if m == "" {
		panic("MASK_SECRET env not found")
	}
	secretKey = []byte(m)
}

func Mask(id uint) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(fmt.Sprintf("%d", id)))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	ctr := cipher.NewCTR(block, iv)
	plaintext := []byte(fmt.Sprintf("%d", id))
	ctr.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func Unmask(ciphertext string) (uint, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return 0, err
	}

	decodedCiphertext, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return 0, err
	}

	iv := decodedCiphertext[:aes.BlockSize]
	ctr := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(decodedCiphertext)-aes.BlockSize)
	ctr.XORKeyStream(plaintext, decodedCiphertext[aes.BlockSize:])

	decryptedID, err := strconv.Atoi(string(plaintext))
	if err != nil {
		return 0, err
	}
	return uint(decryptedID), nil
}
