package store

import "errors"

type BytesRef struct {
	bs     []byte
	offset int
	length int
}

func NewBytesRef(bs []byte) *BytesRef {
	return &BytesRef{
		bs:     bs,
		offset: 0,
		length: 0,
	}
}

func NewMustBytesRef(bs []byte, offset, length int) *BytesRef {
	return &BytesRef{
		bs:     bs,
		offset: offset,
		length: length,
	}
}

func (b *BytesRef) SetOffset(offset int) error {
	if offset < 0 || offset > len(b.bs) {
		return errors.New("")
	}
	b.offset = offset
	return nil
}

func (b *BytesRef) SetLength(length int) error {
	if len(b.bs[b.offset:]) < length {
		return errors.New("")
	}
	b.length = length
	return nil
}

func (b *BytesRef) Set(offset, length int) error {
	if offset < 0 || offset > len(b.bs) {
		return errors.New("")
	}
	if len(b.bs[offset:]) < length {
		return errors.New("")
	}

	b.offset = offset
	b.length = length

	return nil
}

func (b *BytesRef) Bytes() []byte {
	return b.bs[b.offset : b.offset+b.length]
}

func (b *BytesRef) RawBytes() []byte {
	return b.bs
}

func (b *BytesRef) Offset() int {
	return b.offset
}

func (b *BytesRef) Len() int {
	return b.length
}
