package store

import (
	"errors"
	"io"
)

var _ IndexInput = &ByteArrayDataInput{}

// ByteArrayDataInput DataInput backed by a byte array. WARNING: This class omits all low-level checks.
type ByteArrayDataInput struct {
	*BaseDataInput

	bs  []byte
	pos int
}

func NewByteArrayDataInput(bs []byte) *ByteArrayDataInput {
	input := &ByteArrayDataInput{
		bs:  bs,
		pos: 0,
	}

	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

func (b *ByteArrayDataInput) Reset(bs []byte) {
	b.buff = bs
	b.pos = 0
}

func (b *ByteArrayDataInput) GetPosition() int {
	return b.pos
}

func (b *ByteArrayDataInput) Seek(offset int64, whence int) (int64, error) {
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

func (b *ByteArrayDataInput) GetFilePointer() int64 {
	return int64(b.pos)
}

func (b *ByteArrayDataInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	bs := b.bs[offset : offset+length]
	return NewByteArrayDataInput(bs), nil
}

func (b *ByteArrayDataInput) Length() int64 {
	return int64(len(b.bs))
}

func (b *ByteArrayDataInput) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	bs := b.bs[offset : offset+length]
	return &randomAccessIndexInput{in: NewByteArrayDataInput(bs)}, nil
}

func (b *ByteArrayDataInput) Read(p []byte) (n int, err error) {
	less := len(b.bs) - b.pos

	copySize := len(p)
	if len(p) > less {
		copySize = less
	}

	copy(p, b.bs[b.pos:b.pos+copySize])
	b.pos += copySize

	return copySize, nil
}

func (b *ByteArrayDataInput) Clone() CloneReader {
	input := &ByteArrayDataInput{
		bs:  b.bs,
		pos: b.pos,
	}

	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

func (b *ByteArrayDataInput) Rewind() {
	b.pos = 0
}

func (b *ByteArrayDataInput) SetPosition(pos int) {
	b.pos = pos
}

var _ DataOutput = &ByteArrayDataOutput{}

// ByteArrayDataOutput DataOutput backed by a byte array.
// WARNING: This class omits most low-level checks, so be sure to test heavily with assertions enabled.
type ByteArrayDataOutput struct {
	*BaseDataOutput

	bytes []byte
	pos   int
}

func NewByteArrayDataOutput(bytes []byte) *ByteArrayDataOutput {
	output := &ByteArrayDataOutput{bytes: bytes}
	output.BaseDataOutput = NewBaseDataOutput(output)
	return output
}

func (r *ByteArrayDataOutput) Write(b []byte) (int, error) {
	if r.pos+len(b) > len(r.bytes) {
		return 0, errors.New("input data too long")
	}

	copy(r.bytes[r.pos:], b)
	r.pos += len(b)
	return len(b), nil
}

func (r *ByteArrayDataOutput) Reset(bytes []byte) error {
	return r.ResetAt(bytes, 0, len(bytes))
}

func (r *ByteArrayDataOutput) ResetAt(bytes []byte, offset, size int) error {
	if offset >= len(bytes) {
		return errors.New("offset over len(bytes)")
	}

	r.bytes = bytes[offset : offset+size]
	r.pos = 0
	return nil
}

func (r *ByteArrayDataOutput) GetPosition() int {
	return r.pos
}
