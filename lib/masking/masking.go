package masking

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
)

// todo: .ENV
var secretKey = []byte("your_16_byte_key")

func Mask(id int) (string, error) {
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

	decryptedStr := string(plaintext)
	decryptedUint, err := strconv.Atoi(decryptedStr)
	if err != nil {
		return 0, err
	}

	return uint(decryptedUint), nil
}
