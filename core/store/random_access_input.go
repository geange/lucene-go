package store

// RandomAccessInput Random Access Index API. Unlike IndexInput, this has no concept of file position,
// all reads are absolute. However, like IndexInput, it is only intended for use by a single thread.
type RandomAccessInput interface {

	// RUint8 Reads a byte at the given position in the file
	// See Also: DataInput.readByte
	RUint8(pos int64) (byte, error)

	// RUint16 Reads a short at the given position in the file
	// See Also: DataInput.readShort
	RUint16(pos int64) (uint16, error)

	// RUint32 Reads an integer at the given position in the file
	// See Also: DataInput.readInt
	RUint32(pos int64) (uint32, error)

	// RUint64 Reads a long at the given position in the file
	// See Also: DataInput.readLong
	RUint64(pos int64) (uint64, error)
}
