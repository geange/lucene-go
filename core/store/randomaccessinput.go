package store

import (
	"encoding/binary"
	"io"
)

// RandomAccessInput Random Access Index API. Unlike IndexInput, this has no concept of file pos,
// all reads are absolute. However, like IndexInput, it is only intended for use by a single thread.
type RandomAccessInput interface {
	io.ReaderAt

	// ReadU8
	// Reads a byte at the given pos in the file
	// See Also: DataInput.readByte
	ReadU8(pos int64) (byte, error)

	// ReadU16
	// Reads a short at the given pos in the file
	// See Also: DataInput.readShort
	ReadU16(pos int64) (uint16, error)

	// ReadU32
	// Reads an integer at the given pos in the file
	// See Also: DataInput.readInt
	ReadU32(pos int64) (uint32, error)

	// ReadU64
	// Reads a long at the given pos in the file
	// See Also: DataInput.readLong
	ReadU64(pos int64) (uint64, error)
}

var _ RandomAccessInput = &randomAccessIndexInput{}

type randomAccessIndexInput struct {
	in IndexInput
}

func (r *randomAccessIndexInput) ReadAt(p []byte, off int64) (int, error) {
	if _, err := r.in.Seek(off, io.SeekStart); err != nil {
		return 0, err
	}
	return r.in.Read(p)
}

func (r *randomAccessIndexInput) ReadU8(pos int64) (byte, error) {
	if _, err := r.in.Seek(pos, io.SeekStart); err != nil {
		return 0, err
	}
	return r.in.ReadByte()
}

func (r *randomAccessIndexInput) ReadU16(pos int64) (uint16, error) {
	if _, err := r.in.Seek(pos, io.SeekStart); err != nil {
		return 0, err
	}
	return r.in.ReadUint16(nil)
}

func (r *randomAccessIndexInput) ReadU32(pos int64) (uint32, error) {
	if _, err := r.in.Seek(pos, io.SeekStart); err != nil {
		return 0, err
	}
	return r.in.ReadUint32(nil)
}

func (r *randomAccessIndexInput) ReadU64(pos int64) (uint64, error) {
	if _, err := r.in.Seek(pos, io.SeekStart); err != nil {
		return 0, err
	}
	return r.in.ReadUint64(nil)
}

var _ RandomAccessInput = &BytesRandomAccessInput{}

func NewBytesRandomAccessInput(bs []byte, byteOrder binary.ByteOrder) RandomAccessInput {
	return &BytesRandomAccessInput{
		bs:        bs,
		byteOrder: byteOrder,
	}
}

type BytesRandomAccessInput struct {
	bs        []byte
	byteOrder binary.ByteOrder
}

func (b *BytesRandomAccessInput) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(b.bs)) {
		return 0, io.ErrUnexpectedEOF
	}

	copySize := len(p)
	n := len(b.bs[off:])
	if n < copySize {
		copySize = n
	}

	copy(p, b.bs[off:off+int64(copySize)])
	return copySize, nil
}

func (b *BytesRandomAccessInput) ReadU8(pos int64) (byte, error) {
	if pos >= int64(len(b.bs)) {
		return 0, io.ErrUnexpectedEOF
	}
	return b.bs[pos], nil
}

func (b *BytesRandomAccessInput) ReadU16(pos int64) (uint16, error) {
	if pos+2 >= int64(len(b.bs)) {
		return 0, io.ErrUnexpectedEOF
	}
	return b.byteOrder.Uint16(b.bs[pos:]), nil
}

func (b *BytesRandomAccessInput) ReadU32(pos int64) (uint32, error) {
	if pos+4 >= int64(len(b.bs)) {
		return 0, io.ErrUnexpectedEOF
	}
	return b.byteOrder.Uint32(b.bs[pos:]), nil
}

func (b *BytesRandomAccessInput) ReadU64(pos int64) (uint64, error) {
	if pos+8 >= int64(len(b.bs)) {
		return 0, io.ErrUnexpectedEOF
	}
	return b.byteOrder.Uint64(b.bs[pos:]), nil
}
