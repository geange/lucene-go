package store

import "bytes"

var _ DataInput = &ByteArrayDataInput{}

// ByteArrayDataInput DataInput backed by a byte array. WARNING: This class omits all low-level checks.
type ByteArrayDataInput struct {
	*Reader

	buf *bytes.Buffer
}

func NewByteArrayDataInput(bs []byte) *ByteArrayDataInput {
	input := &ByteArrayDataInput{
		buf: bytes.NewBuffer(bs),
	}

	input.Reader = NewReader(input)
	return input
}

func (b *ByteArrayDataInput) ReadByte() (byte, error) {
	return b.buf.ReadByte()
}

func (b *ByteArrayDataInput) Read(p []byte) (n int, err error) {
	return b.buf.Read(p)
}

func (b *ByteArrayDataInput) Close() error {
	//TODO implement me
	panic("implement me")
}
