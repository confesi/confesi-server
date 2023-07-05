package cipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	email string
	id    string
	nonce []byte
}

func (user *testUser) Key() []byte {
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
	return &testUser{"foo@bar.com", "foobarbaz", nil}
}

func testInitWithData() (*testUser, *KDF) {
	kdf, _ := NewKDF()
	user := testInit()
	return user, kdf
}

func TestKDFKeyGen(t *testing.T) {
	kdf, err := NewKDF()
	assert.Nil(t, err)

	// key gen test
	// same key for 1 struct
	user := testInit()

	key1, err := kdf.GenKey(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, key1)
	assert.Equal(t, 32, len(key1))

	key2, err := kdf.GenKey(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, key2)
	assert.Equal(t, 32, len(key2))

	assert.Equal(t, key1, key2)
}

func TestMaskingUserID(t *testing.T) {
	var err error
	user, kdf := testInitWithData()

	key, err := kdf.GenKey(user)
	assert.Nil(t, err)

	// test ciphering
	cipher, err := Cipher([]byte(user.id), key)
	assert.Nil(t, err)
	assert.NotEmpty(t, cipher.Nonce)
	assert.NotEqual(t, string(cipher.Ciphertext), user.id)

	// test deciphering
	err = cipher.Decipher(key)
	assert.Nil(t, err)
	assert.Equal(t, user.id, string(cipher.Ciphertext))
}
