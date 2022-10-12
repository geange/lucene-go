package fst

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/pkg/errors"
)

var _ FSTStore = &OffHeapFSTStore{}

// OffHeapFSTStore Provides off heap storage of finite state machine (FST), using underlying index
// input instead of byte store on heap
// lucene.experimental
type OffHeapFSTStore struct {
	in       store.IndexInput
	offset   int
	numBytes int
}

func (o *OffHeapFSTStore) Init(in store.DataInput, numBytes int64) error {
	input, ok := in.(store.IndexInput)
	if ok {
		o.in = input
		o.numBytes = int(numBytes)
		o.offset = int(o.in.GetFilePointer())
		return nil
	}
	return errors.Wrap(ErrIllegalArgument,
		"in should be an instance of IndexInput")
}

func (o *OffHeapFSTStore) Size() int64 {
	return int64(o.numBytes)
}

func (o *OffHeapFSTStore) GetReverseBytesReader() BytesReader {
	slice, err := o.in.RandomAccessSlice(int64(o.offset), int64(o.numBytes))
	if err != nil {
		return nil
	}
	return NewReverseRandomAccessReader(slice)
}

func (o *OffHeapFSTStore) WriteTo(out store.DataOutput) error {
	return errors.Wrap(ErrUnsupportedOperation,
		"writeToOutput operation is not supported for OffHeapFSTStore")
}
