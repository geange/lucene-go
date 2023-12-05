package fst

import (
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"io"
)

var _ Store = &OnHeapFSTStore{}

// OnHeapFSTStore Provides storage of finite state machine (FST),
// using byte array or byte store allocated on heap.
// lucene.experimental
type OnHeapFSTStore struct {
	bytes        *ByteStore
	bytesArray   []byte
	maxBlockBits int
}

func NewOnHeapFSTStore(maxBlockBits int) (*OnHeapFSTStore, error) {
	if maxBlockBits < 1 || maxBlockBits > 30 {
		return nil, fmt.Errorf("maxBlockBits should be 1 .. 30; got %d", maxBlockBits)
	}

	return &OnHeapFSTStore{maxBlockBits: maxBlockBits}, nil
}

func (o *OnHeapFSTStore) Init(in io.Reader, numBytes int64) error {
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
	_, err := in.Read(o.bytesArray)
	return err
}

func (o *OnHeapFSTStore) Size() int64 {
	if o.bytesArray != nil {
		return int64(len(o.bytesArray))
	} else {
		return 0
	}
}

func (o *OnHeapFSTStore) GetReverseBytesReader() (BytesReader, error) {
	if o.bytesArray != nil {
		return NewReverseBytesReader(o.bytesArray), nil
	} else {
		return o.bytes.GetReverseReader()
	}
}

func (o *OnHeapFSTStore) WriteTo(out store.DataOutput) error {
	if o.bytes != nil {
		numBytes := o.bytes.GetPosition()
		if err := out.WriteUvarint(uint64(numBytes)); err != nil {
			return err
		}
		return o.bytes.WriteTo(out)
	}

	// TODO: assert bytesArray != null;
	if err := out.WriteUvarint(uint64(len(o.bytesArray))); err != nil {
		return err
	}
	_, err := out.Write(o.bytesArray)
	return err
}
