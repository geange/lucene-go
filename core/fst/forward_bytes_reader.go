package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &ForwardBytesReader{}

// ForwardBytesReader Reads from a single byte[].
type ForwardBytesReader struct {
	bytes []byte
	pos   int
}

func (f *ForwardBytesReader) WriteByte(b byte) error {
	//TODO implement me
	panic("implement me")
}

func (f *ForwardBytesReader) WriteBytes(b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (f *ForwardBytesReader) CopyBytes(input store.DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (f *ForwardBytesReader) GetPosition() int64 {
	//TODO implement me
	panic("implement me")
}

func (f *ForwardBytesReader) SetPosition(pos int64) error {
	//TODO implement me
	panic("implement me")
}

func (f *ForwardBytesReader) Reversed() bool {
	//TODO implement me
	panic("implement me")
}
