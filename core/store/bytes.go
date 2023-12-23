package store

import (
	"errors"
	"io"
)

var _ IndexInput = &BytesInput{}

// BytesInput DataInput backed by a byte array. WARNING: This class omits all low-level checks.
type BytesInput struct {
	*BaseDataInput

	bs  []byte
	pos int
}

func (b *BytesInput) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		b.pos = int(offset)
	case io.SeekCurrent:
		b.pos += int(offset)
	case io.SeekEnd:
		b.pos = len(b.bs) - int(offset) - 1
	}
	return int64(b.pos), nil
}

func (b *BytesInput) GetFilePointer() int64 {
	return int64(b.pos)
}

func (b *BytesInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	bs := b.bs[offset : offset+length]
	return NewBytesInput(bs), nil
}

func (b *BytesInput) Length() int64 {
	return int64(len(b.bs))
}

func (b *BytesInput) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	bs := b.bs[offset : offset+length]
	return &randomAccessIndexInput{in: NewBytesInput(bs)}, nil
}

func NewBytesInput(bs []byte) *BytesInput {
	input := &BytesInput{
		bs:  bs,
		pos: 0,
	}

	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

func (b *BytesInput) Read(p []byte) (n int, err error) {
	less := len(b.bs) - b.pos

	copySize := len(p)
	if len(p) > less {
		copySize = less
	}

	copy(p, b.bs[b.pos:b.pos+copySize])
	b.pos += copySize

	return copySize, nil
}

func (b *BytesInput) Clone() CloneReader {
	input := &BytesInput{
		bs:  b.bs,
		pos: b.pos,
	}

	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

var _ DataOutput = &BytesOutput{}

// BytesOutput DataOutput backed by a byte array. WARNING: This class omits most low-level checks, so be sure to test heavily with assertions enabled.
type BytesOutput struct {
	*BaseDataOutput

	bytes []byte
	pos   int
}

func NewBytesDataOutput(bytes []byte) *BytesOutput {
	output := &BytesOutput{bytes: bytes}
	output.BaseDataOutput = NewBaseDataOutput(output)
	return output
}

func (r *BytesOutput) Write(b []byte) (int, error) {
	if r.pos+len(b) > len(r.bytes) {
		return 0, errors.New("input data too long")
	}

	copy(r.bytes[r.pos:], b)
	r.pos += len(b)
	return len(b), nil
}

func (r *BytesOutput) Reset(bytes []byte) error {
	return r.ResetAt(bytes, 0, len(bytes))
}

func (r *BytesOutput) ResetAt(bytes []byte, offset, size int) error {
	if offset >= len(bytes) {
		return errors.New("offset over len(bytes)")
	}

	r.bytes = bytes[offset : offset+size]
	r.pos = 0
	return nil
}

func (r *BytesOutput) GetPosition() int {
	return r.pos
}
