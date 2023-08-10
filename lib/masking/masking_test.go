package masking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionAndDecryption(t *testing.T) {
	tests := []struct {
		id int
	}{
		{0},      // test case 1
		{12345},  // test case 2
		{987654}, // test case 3
		{-42},    // test case 4
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

	assert.Equal(t, 123, decrypted)
}
