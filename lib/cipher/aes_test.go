package cipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskingUserID(t *testing.T) {
	var err error
	user, kdf := testInitWithData()

	// copy user id
	userid := make([]byte, len(user.id))
	copy(userid, []byte(user.id))

	key, err := kdf.GenKey(user)
	assert.Nil(t, err)

	// test ciphering
	err = Cipher(user, key)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.nonce)
	assert.NotEqual(t, string(userid), user.id)

	// test deciphering
	err = Decipher(user, key)
	assert.Nil(t, err)
	assert.Equal(t, user.id, string(userid))
}
