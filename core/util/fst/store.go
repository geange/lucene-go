package fst

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/geange/lucene-go/core/store"
)

// Store Abstraction for reading/writing bytes necessary for FST.
type Store interface {
	Init(in io.Reader, numBytes int64) error

	Size() int64

	GetReverseBytesReader() (BytesReader, error)

	WriteTo(ctx context.Context, out store.DataOutput) error
}

var _ Store = &OnHeapStore{}

// OnHeapStore Provides storage of finite state machine (FST),
// using byte array or byte store allocated on heap.
type OnHeapStore struct {
	bytes        *ByteStore
	bytesArray   []byte
	maxBlockBits int
}

func NewOnHeapStore(maxBlockBits int) (*OnHeapStore, error) {
	if maxBlockBits < 1 || maxBlockBits > 30 {
		return nil, fmt.Errorf("maxBlockBits should be 1 .. 30; got %d", maxBlockBits)
	}

	return &OnHeapStore{maxBlockBits: maxBlockBits}, nil
}

func (o *OnHeapStore) Init(in io.Reader, numBytes int64) error {
	if numBytes > 1<<o.maxBlockBits {
		// Fst is big: we need multiple pages
		bytes, err := NewBytesStoreByDataInput(in, numBytes, 1<<o.maxBlockBits)
		if err != nil {
			return err
		}
		o.bytes = bytes
		return nil
	}

	// Fst fits into a single block: use ByteArrayBytesStoreReader for less overhead
	o.bytesArray = make([]byte, numBytes)
	if _, err := in.Read(o.bytesArray); err != nil {
		return err
	}
	return nil
}

func (o *OnHeapStore) Size() int64 {
	if o.bytesArray != nil {
		return int64(len(o.bytesArray))
	}
	return 0
}

func (o *OnHeapStore) GetReverseBytesReader() (BytesReader, error) {
	if o.bytesArray != nil {
		return newReverseBytesReader(o.bytesArray), nil
	}
	return o.bytes.GetReverseReader()
}

func (o *OnHeapStore) WriteTo(ctx context.Context, out store.DataOutput) error {
	if o.bytes != nil {
		numBytes := o.bytes.GetPosition()
		if err := out.WriteUvarint(ctx, uint64(numBytes)); err != nil {
			return err
		}
		return o.bytes.WriteToDataOutput(out)
	}

	if err := out.WriteUvarint(ctx, uint64(len(o.bytesArray))); err != nil {
		return err
	}
	if _, err := out.Write(o.bytesArray); err != nil {
		return err
	}
	return nil
}

// OffHeapStore Provides off heap storage of finite state machine (FST),
// using underlying index input instead of byte store on heap
type OffHeapStore struct {
	in       store.IndexInput
	offset   int64
	numBytes int64
}

func (o *OffHeapStore) Init(r io.Reader, numBytes int64) error {
	if in, ok := r.(store.IndexInput); ok {
		o.in = in
		o.numBytes = numBytes
		o.offset = in.GetFilePointer()
		return nil
	}
	return errors.New("parameter:in should be an instance of IndexInput for using OffHeapFSTStore")
}

func (o *OffHeapStore) Size() int64 {
	return o.numBytes
}

func (o *OffHeapStore) GetReverseBytesReader() (BytesReader, error) {
	input, err := o.in.RandomAccessSlice(o.offset, o.numBytes)
	if err != nil {
		return nil, err
	}
	return newReverseRandomAccessReader(input), nil
}

func (o *OffHeapStore) WriteTo(ctx context.Context, out store.DataOutput) error {
	return errors.New("writeToOutput operation is not supported for OffHeapFSTStore")
}
