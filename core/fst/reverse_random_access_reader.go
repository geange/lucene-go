package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &ReverseRandomAccessReader{}

// ReverseRandomAccessReader Implements reverse read from a RandomAccessInput.
type ReverseRandomAccessReader struct {
	in  store.RandomAccessInput
	pos int64
}

func (r *ReverseRandomAccessReader) WriteByte(b byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseRandomAccessReader) WriteBytes(b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseRandomAccessReader) CopyBytes(input store.DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseRandomAccessReader) GetPosition() int64 {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseRandomAccessReader) SetPosition(pos int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *ReverseRandomAccessReader) Reversed() bool {
	//TODO implement me
	panic("implement me")
}
