package masking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueMasksMapToSameID(t *testing.T) {
	// Test case: Masking the same ID twice should result in different encrypted values
	id := uint(5)

	maskedID1, err := Mask(id)
	assert.NoError(t, err, "Masking error")

	maskedID2, err := Mask(id)
	assert.NoError(t, err, "Masking error")

	assert.NotEqual(t, maskedID1, maskedID2, "Masked IDs should not be equal")

	// Test case: Unmasking the encrypted IDs should yield the same original ID
	decryptedID1, err := Unmask(maskedID1)
	assert.NoError(t, err, "Unmasking error")

	decryptedID2, err := Unmask(maskedID2)
	assert.NoError(t, err, "Unmasking error")

	assert.Equal(t, id, decryptedID1, "Original and decrypted IDs do not match")
	assert.Equal(t, id, decryptedID2, "Original and decrypted IDs do not match")
}

func TestEncryptionAndDecryption(t *testing.T) {
	tests := []struct {
		id uint
	}{
		{0},                 // sub-test case 1
		{12345212121224583}, // sub-test case 2
		{987654},            // sub-test case 3
		{42},                // sub-test case 4
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ID_%d", test.id), func(t *testing.T) {
			encrypted, err := Mask(test.id)
			if err != nil {
				t.Errorf("Encryption error: %v", err)
			}

			decrypted, err := Unmask(encrypted)
			if err != nil {
				t.Errorf("Decryption error: %v", err)
			}

			assert.Equal(t, test.id, decrypted, "Original and decrypted IDs do not match")
		})
	}
}

func TestEncryptionAndDecryptionSimple(t *testing.T) {
	val, err := Mask(123)
	if err != nil {
		t.Error("Encryption error:", err)
	}

	decrypted, err := Unmask(val)
	if err != nil {
		t.Error("Decryption error:", err)
	}

	assert.Equal(t, uint(123), decrypted)
}
