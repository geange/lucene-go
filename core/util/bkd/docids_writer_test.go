package bkd

import (
	"math/rand"
	"testing"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestWriteDocIdsSorted(t *testing.T) {
	output := store.NewByteBuffersDataOutput()
	docIds := make([]int, 100)
	for i := range docIds {
		docIds[i] = i
	}
	err := WriteDocIds(docIds, output)
	assert.Nil(t, err)

	input := store.NewByteArrayDataInput(output.Bytes())

	newDocIds := make([]int, 100)
	err = ReadInts(input, 100, newDocIds)
	assert.Nil(t, err)

	assert.Equal(t, docIds, newDocIds)
}

func TestWriteDocIdsInt24(t *testing.T) {
	output := store.NewByteBuffersDataOutput()
	docIds := make([]int, 100)
	for i := range docIds {
		docIds[i] = rand.Intn(0xFFFFFF)
	}
	err := WriteDocIds(docIds, output)
	assert.Nil(t, err)

	input := store.NewByteArrayDataInput(output.Bytes())

	newDocIds := make([]int, 100)
	err = ReadInts(input, 100, newDocIds)
	assert.Nil(t, err)

	assert.Equal(t, docIds, newDocIds)
}

func TestWriteDocIdsInt32(t *testing.T) {
	output := store.NewByteBuffersDataOutput()
	docIds := make([]int, 100)
	for i := range docIds {
		docIds[i] = rand.Intn(0xFFFFFF) + 0xFFFFFF
	}
	err := WriteDocIds(docIds, output)
	assert.Nil(t, err)

	input := store.NewByteArrayDataInput(output.Bytes())

	newDocIds := make([]int, 100)
	err = ReadInts(input, 100, newDocIds)
	assert.Nil(t, err)

	assert.Equal(t, docIds, newDocIds)
}
