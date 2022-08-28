package fst

import "github.com/geange/lucene-go/core/store"

var _ FSTStore = &OnHeapFSTStore{}

// OnHeapFSTStore Provides storage of finite state machine (FST), using byte array or byte store allocated on heap.
// lucene.experimental
type OnHeapFSTStore struct {
	bytes        *BytesStore
	bytesArray   []byte
	maxBlockBits int
}

func NewOnHeapFSTStore(maxBlockBits int) *OnHeapFSTStore {
	return &OnHeapFSTStore{maxBlockBits: maxBlockBits}
}

func (r *OnHeapFSTStore) Init(in store.DataInput, numBytes int64) error {
	if numBytes > 1<<r.maxBlockBits {
		var err error
		r.bytes, err = NewBytesStore3(in, int(numBytes), 1<<r.maxBlockBits)
		if err != nil {
			return err
		}
	} else {
		r.bytesArray = make([]byte, int(numBytes))
		return in.ReadBytes(r.bytesArray)
	}
	return nil
}

func (r *OnHeapFSTStore) Size() int64 {
	if r.bytesArray != nil {
		return int64(len(r.bytesArray))
	}
	return 0
}

func (r *OnHeapFSTStore) GetReverseBytesReader() BytesReader {
	if r.bytesArray != nil {
		return NewReverseBytesReader(r.bytesArray)
	}
	return r.bytes.GetReverseReader()
}

func (r *OnHeapFSTStore) WriteTo(out store.DataOutput) error {
	if r.bytes != nil {
		numBytes := r.bytes.GetPosition()
		err := out.WriteUvarint(uint64(numBytes))
		if err != nil {
			return err
		}
	} else {
		err := out.WriteUvarint(uint64(len(r.bytesArray)))
		if err != nil {
			return err
		}
		return out.WriteBytes(r.bytesArray)
	}
	return nil
}
