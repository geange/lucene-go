package fst

import "github.com/geange/lucene-go/core/store"

// FSTStore Abstraction for reading/writing bytes necessary for FST.
type FSTStore interface {
	Init(in store.DataInput, numBytes int64) error

	Size() int64

	GetReverseBytesReader() BytesReader

	WriteTo(out store.DataOutput) error
}
