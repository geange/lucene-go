package fst

import (
	"github.com/geange/lucene-go/core/store"
	"io"
)

// Store Abstraction for reading/writing bytes necessary for FST.
type Store interface {
	Init(in io.Reader, numBytes int64) error

	Size() int64

	GetReverseBytesReader() (BytesReader, error)

	WriteTo(out store.DataOutput) error
}
