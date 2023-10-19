package bkd

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestNewHeapPointWriter(t *testing.T) {
	config, err := getRandomConfig()
	assert.Nil(t, err)

	size := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(15000)

	writer := NewHeapPointWriter(config, size)
	for docId := 0; docId < size; docId++ {
		packedValue := getPackedValue(config)
		err := writer.Append(packedValue, docId)
		assert.Nil(t, err)
	}
}
