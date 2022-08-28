package fst

import (
	"io"

	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
)

var _ BytesReader = &ForwardBytesReader{}

// TODO: can we use just ByteArrayDataInput...?  need to
// add a .skipBytes to DataInput.. hmm and .setPosition

// ForwardBytesReader Reads from a single byte[].
type ForwardBytesReader struct {
	*store.DataInputImp

	bytes []byte
	pos   int

	isEOF bool
}

func NewForwardBytesReader(bytes []byte) *ForwardBytesReader {
	input := &ForwardBytesReader{bytes: bytes}
	input.DataInputImp = store.NewDataInputImp(input)
	return input
}

func (f *ForwardBytesReader) ReadByte() (byte, error) {
	if f.isEOF {
		return 0, io.EOF
	}

	v := f.bytes[f.pos]
	f.pos++

	if int(f.pos) == len(f.bytes) {
		f.isEOF = true
	}

	return v, nil
}

func (f *ForwardBytesReader) ReadBytes(b []byte) error {
	if f.isEOF {
		return io.EOF
	}

	copy(f.bytes[f.pos:], b)

	f.pos += Max(len(f.bytes[f.pos:]), len(b))
	if int(f.pos) == len(f.bytes) {
		f.isEOF = true
	}
	return nil
}

func (f *ForwardBytesReader) SkipBytes(count int) error {
	f.pos += count
	if f.pos >= len(f.bytes) {
		f.pos = len(f.bytes)
		f.isEOF = true
	}
	return nil
}

func (f *ForwardBytesReader) GetPosition() int {
	return f.pos
}

func (f *ForwardBytesReader) SetPosition(pos int) {
	if pos > len(f.bytes) {
		f.pos = len(f.bytes)
		f.isEOF = true
		return
	}
	f.pos = int(pos)
}

func (f *ForwardBytesReader) Reversed() bool {
	return false
}
