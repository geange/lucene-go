package fst

import (
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/geange/lucene-go/core/store"
)

var _ store.DataOutput = &BytesStore{}

// BytesStore
// TODO: merge with PagedBytes, except PagedBytes doesn't
// let you read while writing which FST needs
type BytesStore struct {
	blocks    *arraylist.List
	blockSize int
	blockBits int
	blockMask int
	current   []byte
	nextWrite int
}

func (b *BytesStore) WriteByte(c byte) error {
	//TODO implement me
	panic("implement me")
}

func (b *BytesStore) WriteBytes(bs []byte) error {
	//TODO implement me
	panic("implement me")
}

func (b *BytesStore) CopyBytes(input store.DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}
