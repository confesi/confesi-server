package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

var block cipher.Block

// Strictness ensures that decode(x) = decode(y) only if x = y (I hope). This isn't targeted at a concrete problem, but is a better default.
var encoding *base64.Encoding = base64.RawURLEncoding.Strict()

func init() {
	// load from .env
	m := os.Getenv("MASK_SECRET")
	if m == "" {
		panic("MASK_SECRET env not found")
	}

	secretKey, err := encoding.DecodeString(m)
	if err != nil {
		panic(fmt.Errorf("couldn't decode MASK_SECRET: %w", err))
	}

	block, err = aes.NewCipher(secretKey)
	if err != nil {
		panic(fmt.Errorf("couldn't use MASK_SECRET as AES key: %w", err))
	}
}

func encrypt(id uint32) string {
	buf := make([]byte, aes.BlockSize)
	binary.LittleEndian.PutUint32(buf[:4], id)
	block.Encrypt(buf, buf)
	return encoding.EncodeToString(buf)
}

func Hash(id uint) string {
	if id > math.MaxUint32 {
		panic("id out of range")
	}

	return encrypt(uint32(id))
}

func Mask(id uint) (string, error) {
	if id > math.MaxUint32 {
		return "", fmt.Errorf("id out of range")
	}

	return encrypt(uint32(id)), nil
}

func Unmask(ciphertext string) (uint, error) {
	if len(ciphertext) != encoding.EncodedLen(aes.BlockSize) {
		return 0, fmt.Errorf("invalid ciphertext length")
	}

	buf, err := encoding.DecodeString(ciphertext)
	if err != nil {
		return 0, err
	}

	block.Decrypt(buf, buf)

	// 256 - 32 = 224 bits for authenticated encryption. This check doesn't need to be timing-safe.
	for _, b := range(buf[4:]) {
		if b != 0 {
			return 0, fmt.Errorf("invalid ciphertext")
		}
	}

	return uint(binary.LittleEndian.Uint32(buf[:4])), nil
}
