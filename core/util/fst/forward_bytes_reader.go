package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &ForwardBytesReader{}

type ForwardBytesReader struct {
	bytes []byte
	pos   int64

	*store.DataInputImp
}

func NewForwardBytesReader(bytes []byte) *ForwardBytesReader {
	reader := &ForwardBytesReader{bytes: bytes}
	reader.DataInputImp = store.NewDataInputImp(reader)
	return reader
}

func (f *ForwardBytesReader) ReadByte() (byte, error) {
	b := f.bytes[f.pos]
	f.pos++
	return b, nil
}

func (f *ForwardBytesReader) ReadBytes(b []byte) error {
	copy(b, f.bytes[f.pos:])
	f.pos += int64(len(b))
	return nil
}

func (f *ForwardBytesReader) GetPosition() int64 {
	return f.pos
}

func (f *ForwardBytesReader) SetPosition(pos int64) {
	f.pos = pos
}

func (f *ForwardBytesReader) Reversed() bool {
	return false
}
