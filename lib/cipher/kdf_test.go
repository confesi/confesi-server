package cipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	email string
	id    string
}

func (user *testUser) Mask(key []byte) {
	user.id = string(key)
}

func (user *testUser) MasterKey() []byte {
	keyLen := len(user.email)
	if keyLen > MasterKeyLen {
		return []byte(user.email)[:MasterKeyLen]
	}

	if keyLen < MasterKeyLen {
		offset := MasterKeyLen - keyLen
		for i := 0; i < offset; i++ {
			user.email += " "
		}
	}
	return []byte(user.email)
}

func testInit() *testUser {
	return &testUser{"foo@bar.com", "foobarbaz"}
}

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
