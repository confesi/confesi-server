package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCipherOps(t *testing.T) {
	testCases := []struct {
		plaintext string
		ad        []byte
	}{
		{"foo", []byte("userid1")},
		{"bar", []byte("userid2")},
		{"baz", []byte("userid3")},
		{"foobarbaz1", []byte("testuserid1")},
		{"foobarbaz2", []byte("testuserid2")},
		{"foobarbaz3", []byte("testuserid3")},
	}

	for _, v := range testCases {
		// test ciphering
		ciphertext, err := Cipher([]byte(v.plaintext), v.ad)
		assert.Nil(t, err)
		assert.NotEqual(t, ciphertext, []byte(v.plaintext))

		// test deciphering
		pt, err := Decipher(ciphertext, v.ad)
		assert.Nil(t, err)
		assert.Equal(t, string(pt), v.plaintext)
	}
}

func TestCipherMissingAD(t *testing.T) {
	pt := "foobar"
	ad := "baz"

	ciphertext, err := Cipher([]byte(pt), []byte(ad))
	assert.Nil(t, err)

	c1, err := Cipher([]byte(pt), []byte(""))
	assert.NotNil(t, err)
	assert.Nil(t, c1)

	pt2, err := Cipher(ciphertext, []byte(""))
	assert.NotNil(t, err)
	assert.NotEqual(t, pt, string(pt2))
}
