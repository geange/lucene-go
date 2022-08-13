package fst

import (
	"github.com/geange/lucene-go/core/store"
)

var _ BytesReader = &ReverseBytesReader{}

// ReverseBytesReader Reads in reverse from a single byte[].
type ReverseBytesReader struct {
	bytes []byte
	pos   int
}

func (r *ReverseBytesReader) WriteByte(b byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseBytesReader) WriteBytes(b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseBytesReader) CopyBytes(input store.DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseBytesReader) GetPosition() int64 {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseBytesReader) SetPosition(pos int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseBytesReader) Reversed() bool {
	//TODO implement me
	panic("implement me")
}
