package cipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKDFKeyGen(t *testing.T) {
	kdf, err := NewKDF()
	assert.Nil(t, err)

	// salt valid
	assert.NotEmpty(t, kdf.Salt())

	// key gen test
	user := testInit()
	key1, err := kdf.GenKey(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, key1)

	key2, err := kdf.GenKey(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, key2)

	assert.Equal(t, key1, key2)
}

func TestKDFWithSalt(t *testing.T) {
	kdf, err := NewKDF()
	assert.Nil(t, err)

	salt := kdf.Salt()
	kdf, err = NewWithSalt(salt)
	assert.Nil(t, err)

	user := testInit()
	key1, err := kdf.GenKey(user)
	assert.Nil(t, err)
	key2, err := kdf.GenKey(user)
	assert.Nil(t, err)

	assert.Equal(t, key1, key2)
}
