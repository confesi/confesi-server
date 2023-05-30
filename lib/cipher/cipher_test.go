package cipher

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCipher(t *testing.T) {
	testStr := "test-user-id"

	encrypted := Encrypt(testStr)
	log.Fatal(encrypted, testStr)
	assert.NotEqual(t, testStr, encrypted)

	decrypted, err := Decrypt(encrypted)
	log.Fatal(decrypted)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testStr, decrypted)
}
