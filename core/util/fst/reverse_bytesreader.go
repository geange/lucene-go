package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &ReverseBytesReader{}

type ReverseBytesReader struct {
	*store.Reader

	bytes []byte
	pos   int64
}

func NewReverseBytesReader(bytes []byte) *ReverseBytesReader {
	reader := &ReverseBytesReader{bytes: bytes}
	reader.Reader = store.NewReader(reader)
	return reader
}

func (r *ReverseBytesReader) ReadByte() (byte, error) {
	b := r.bytes[r.pos]
	r.pos--
	return b, nil
}

func (r *ReverseBytesReader) Read(b []byte) (int, error) {
	for i := 0; i < len(b); i++ {
		b[i] = r.bytes[r.pos]
		r.pos--
	}
	return len(b), nil
}

func (r *ReverseBytesReader) SkipBytes(numBytes int) error {
	r.pos -= int64(numBytes)
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
