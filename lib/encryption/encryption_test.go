package encryption

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

//! Tests require `MASK_SECRET` env var to be set to pass

func TestUniqueHash(t *testing.T) {

	firstMask1, err := Mask(1)
	if err != nil {
		t.Errorf("Encryption error: %v", err)
	}
	secondMask1, err := Mask(1)
	if err != nil {
		t.Errorf("Encryption error: %v", err)
	}

	mask2, err := Mask(2)
	if err != nil {
		t.Errorf("Encryption error: %v", err)
	}

	assert.Equal(t, firstMask1, secondMask1, "Mask should be deterministic")
	assert.NotEqual(t, firstMask1, mask2, "Mask should be unique")
}

func TestEncryptionAndDecryption(t *testing.T) {
	tests := []struct {
		id uint
	}{
		{0},         // sub-test case 1
		{123452121}, // sub-test case 2
		{987654},    // sub-test case 3
		{42},        // sub-test case 4
		{123},       // sub-test case 5
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
