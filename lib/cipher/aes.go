package cipher

import (
	"crypto/aes"
	c "crypto/cipher"
	"crypto/rand"
	"io"
)

func Cipher(d Serializer, key []byte) error {
	if len(key) != 32 {
		return ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	aesGcm, err := c.NewGCM(block)
	if err != nil {
		return err
	}

	cipher := aesGcm.Seal(nil, nonce, d.Mask(), nil)
	d.Serialize(cipher, nonce)

	return nil
}

func Decipher(d Serializer, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesGcm, err := c.NewGCM(block)
	if err != nil {
		return err
	}

	cipher, nonce := d.Deserialize()
	plainText, err := aesGcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return err
	}

	d.Serialize(plainText, nonce)

	return nil
}
