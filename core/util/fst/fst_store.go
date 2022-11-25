package fst

import (
	"github.com/geange/lucene-go/core/store"
)

// Store Abstraction for reading/writing bytes necessary for FST.
type Store interface {
	Init(in store.DataInput, numBytes int64) error

	Size() int64

	GetReverseBytesReader() (BytesReader, error)

	WriteTo(out store.DataOutput) error
}
