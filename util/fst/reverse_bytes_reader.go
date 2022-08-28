package fst

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
	"io"
)

var _ BytesReader = &ReverseBytesReader{}

// ReverseBytesReader Reads in reverse from a single byte[].
type ReverseBytesReader struct {
	*store.DataInputImp

	bytes []byte
	pos   int

	isEOF bool
}

func NewReverseBytesReader(bytes []byte) *ReverseBytesReader {
	input := &ReverseBytesReader{
		bytes: bytes,
		pos:   len(bytes) - 1,
	}
	input.DataInputImp = store.NewDataInputImp(input)
	return input
}

func (r *ReverseBytesReader) ReadByte() (byte, error) {
	if r.isEOF {
		return 0, io.EOF
	}

	v := r.bytes[r.pos]
	r.pos--
	if r.pos < 0 {
		r.isEOF = true
	}
	return v, nil
}

func (r *ReverseBytesReader) ReadBytes(b []byte) error {
	if r.isEOF {
		return io.EOF
	}
	for i := 0; i < len(b); i++ {
		v, err := r.ReadByte()
		if errors.Is(err, io.EOF) {
			return nil
		}
		b[i] = v
	}
	return nil
}

func (r *ReverseBytesReader) SkipBytes(count int) error {
	r.pos -= count
	if r.pos < 0 {
		r.pos = -1
		r.isEOF = true
	}
	return nil
}

func (r *ReverseBytesReader) GetPosition() int {
	return r.pos
}

func (r *ReverseBytesReader) SetPosition(pos int) {
	if pos >= len(r.bytes) {
		r.pos = len(r.bytes)
	} else {
		r.pos = int(pos)
	}
}

func (r *ReverseBytesReader) Reversed() bool {
	return true
}
