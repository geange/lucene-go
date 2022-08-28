package fst

import (
	"errors"
	"io"

	"github.com/geange/lucene-go/core/store"
)

var _ BytesReader = &ReverseRandomAccessReader{}

// ReverseRandomAccessReader Implements reverse read from a RandomAccessInput.
type ReverseRandomAccessReader struct {
	*store.DataInputImp

	in  store.RandomAccessInput
	pos int64

	isEOF bool
}

func (r *ReverseRandomAccessReader) ReadByte() (byte, error) {
	v, err := r.in.ReadUint8(r.pos)
	if err != nil {
		if errors.Is(err, io.EOF) {
			r.isEOF = true
		}
		return 0, err
	}
	r.pos--
	return v, nil
}

func (r *ReverseRandomAccessReader) ReadBytes(b []byte) error {
	if r.isEOF {
		return io.EOF
	}

	for i := 0; i < len(b); i++ {
		v, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		b[i] = v
	}
	return nil
}

func (r *ReverseRandomAccessReader) SkipBytes(count int) error {
	r.pos += int64(count)
	return nil
}

func (r *ReverseRandomAccessReader) GetPosition() int64 {
	return r.pos
}

func (r *ReverseRandomAccessReader) SetPosition(pos int64) {
	r.pos = pos
}

func (r *ReverseRandomAccessReader) Reversed() bool {
	return true
}
