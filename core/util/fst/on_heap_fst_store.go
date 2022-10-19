package fst

import "github.com/geange/lucene-go/core/store"

var _ FSTStore = &OnHeapFSTStore{}

type OnHeapFSTStore struct {
	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB).
	// If the FST is less than 1 GB then bytesArray is set instead.
	bytes *BytesStore

	// Used at read time when the FST fits into a single byte[].
	bytesArray []byte

	maxBlockBits int
}

func NewOnHeapFSTStore(maxBlockBits int) *OnHeapFSTStore {
	return &OnHeapFSTStore{maxBlockBits: maxBlockBits}
}

func (r *OnHeapFSTStore) Init(in store.DataInput, numBytes int64) error {
	if numBytes > 1<<r.maxBlockBits {
		// FST is big: we need multiple pages
		v, err := NewBytesStoreV1(in, numBytes, 1<<r.maxBlockBits)
		if err != nil {
			return err
		}
		r.bytes = v
	} else {
		// FST fits into a single block: use ByteArrayBytesStoreReader for less overhead
		r.bytesArray = make([]byte, numBytes)
		err := in.ReadBytes(r.bytesArray)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *OnHeapFSTStore) Size() int64 {
	if r.bytesArray != nil {
		return int64(len(r.bytesArray))
	}
	return r.bytes.RamBytesUsed()
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
		return r.bytes.WriteTo(out)
	}

	err := assert(r.bytesArray != nil)
	if err != nil {
		return err
	}
	err = out.WriteUvarint(uint64(len(r.bytesArray)))
	if err != nil {
		return err
	}
	return out.WriteBytes(r.bytesArray)
}
