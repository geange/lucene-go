package fst

import "io"

// FSTStore Abstraction for reading/writing bytes necessary for FST.
type FSTStore interface {
	Init(in io.Reader, numBytes int64) error

	Size() int64

	GetReverseBytesReader() BytesReader

	WriteTo(out io.Writer) (int64, error)
}
