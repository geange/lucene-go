package store

// RandomAccessInput Random Access Index API. Unlike IndexInput, this has no concept of file position,
// all reads are absolute. However, like IndexInput, it is only intended for use by a single thread.
type RandomAccessInput interface {

	// ReadUint8 Reads a byte at the given position in the file
	// See Also: DataInput.readByte
	ReadUint8(pos int64) (byte, error)

	// ReadUint16 Reads a short at the given position in the file
	// See Also: DataInput.readShort
	ReadUint16(pos int64) (uint16, error)

	// ReadUint32 Reads an integer at the given position in the file
	// See Also: DataInput.readInt
	ReadUint32(pos int64) (uint32, error)

	// ReadUint64 Reads a long at the given position in the file
	// See Also: DataInput.readLong
	ReadUint64(pos int64) (uint64, error)
}
