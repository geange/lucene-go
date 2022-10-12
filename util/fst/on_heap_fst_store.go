package fst

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
)

var _ FSTStore = &OnHeapFSTStore{}

// OnHeapFSTStore Provides storage of finite state machine (FST), using byte array or byte store allocated on heap.
// lucene.experimental
type OnHeapFSTStore struct {

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB). If the FST is less than 1 GB then bytesArray is set instead.
	bytes *BytesStore

	// Used at read time when the FST fits into a single byte[].
	bytesArray []byte

	maxBlockBits int
}

func NewOnHeapFSTStore(maxBlockBits int) (*OnHeapFSTStore, error) {
	if maxBlockBits < 1 || maxBlockBits > 30 {
		return nil, errors.New("maxBlockBits should be 1 .. 30")
	}
	return &OnHeapFSTStore{maxBlockBits: maxBlockBits}, nil
}

func (o *OnHeapFSTStore) Init(in store.DataInput, numBytes int64) error {
	if numBytes > 1<<o.maxBlockBits {
		bs, err := NewBytesStoreFromDataInput(in, int(numBytes), 1<<o.maxBlockBits)
		if err != nil {
			return err
		}
		o.bytes = bs
		return nil
	}

	o.bytesArray = make([]byte, int(numBytes))
	return in.ReadBytes(o.bytesArray)
}

func (o *OnHeapFSTStore) Size() int64 {
	if o.bytesArray != nil {
		return int64(len(o.bytesArray))
	}

	return 0
}

func (o *OnHeapFSTStore) GetReverseBytesReader() BytesReader {
	if o.bytesArray != nil {
		return NewReverseBytesReader(o.bytesArray)
	}
	return o.bytes.GetReverseReader()
}

func (o *OnHeapFSTStore) WriteTo(out store.DataOutput) error {
	if o.bytes != nil {
		numBytes := o.bytes.getPosition()
		err := out.WriteUvarint(uint64(numBytes))
		if err != nil {
			return err
		}
		return o.bytes.WriteTo(out)
	}

	err := out.WriteUvarint(uint64(len(o.bytesArray)))
	if err != nil {
		return err
	}
	return out.WriteBytes(o.bytesArray)
}
