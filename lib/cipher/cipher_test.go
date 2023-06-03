package cipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCipher(t *testing.T) {
	testStr := "test-user-input"

	// encryption test
	encrypted := Encrypt(testStr)
	assert.NotEqual(t, testStr, encrypted)

	// description test
	decrypted, err := Decrypt(encrypted)
	assert.Nil(t, err)
	assert.Equal(t, testStr, decrypted)
}
