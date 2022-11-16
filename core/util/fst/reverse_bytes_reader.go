package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &ReverseBytesReader{}

type ReverseBytesReader struct {
	*store.DataInputImp

	bytes []byte
	pos   int64
}

func NewReverseBytesReader(bytes []byte) *ReverseBytesReader {
	reader := &ReverseBytesReader{
		bytes: bytes}
	reader.DataInputImp = store.NewDataInputImp(reader)
	return reader
}

func (r *ReverseBytesReader) ReadByte() (byte, error) {
	b := r.bytes[r.pos]
	r.pos--
	return b, nil
}

func (r *ReverseBytesReader) ReadBytes(b []byte) error {
	for i := 0; i < len(b); i++ {
		b[i] = r.bytes[r.pos]
		r.pos--
	}
	return nil
}

func (r *ReverseBytesReader) GetPosition() int64 {
	return r.pos
}

func (r *ReverseBytesReader) SetPosition(pos int64) error {
	r.pos = pos
	return nil
}

func (r *ReverseBytesReader) Reversed() bool {
	return true
}
