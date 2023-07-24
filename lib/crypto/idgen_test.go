package crypto

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCryptoGenID(t *testing.T) {
	t.Parallel()
	go RefreshCounterMap()

	iter := math.MaxUint16
	userID := "e0640eb6-29cd-11ee-be56-0242ac120002"
	idSet := make(map[string]bool)

	for i := 0; i < iter; i++ {
		id, err := NewID(userID)
		assert.Nil(t, err)
		idSet[id] = true
	}
	assert.Equal(t, iter, len(idSet))

	time.Sleep(time.Second)

	for i := 0; i < iter; i++ {
		id, err := NewID(userID)
		assert.Nil(t, err)
		idSet[id] = true
	}
	assert.Equal(t, iter*2, len(idSet))
}
