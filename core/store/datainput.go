package store

import (
	"io"
)

const (
	SKIP_BUFFER_SIZE = 1024
)

// DataInput Abstract base class for performing read operations of Lucene's low-level data types.
// DataInput may only be used from one thread, because it is not thread safe (it keeps internal state
// like file pos). To allow multithreaded use, every DataInput instance must be cloned before used
// in another thread. Subclasses must therefore implement clone(), returning a new DataInput which operates
// on the same underlying resource, but positioned independently.
type DataInput interface {
	// ByteReader Reads and returns a single byte.
	// See Also: DataOutput.writeByte(byte)
	//ReadByte() (byte, error)
	io.ByteReader

	// Reader Reads a specified number of bytes into an array.
	//ReadBytes(b []byte) error
	io.Reader

	// ReadUint16 Reads two bytes and returns a short.
	// See Also: DataOutput.writeByte(byte)
	ReadUint16() (uint16, error)

	// ReadUint32 Reads four bytes and returns an int.
	// See Also: DataOutput.writeInt(int)
	ReadUint32() (uint32, error)

	// ReadUvarint Reads an int stored in variable-length format. Reads between one and five bytes.
	// Smaller values take fewer bytes. Negative numbers are supported, but should be avoided.
	// The format is described further in DataOutput.writeVInt(int).
	// See Also: DataOutput.writeVInt(int)
	ReadUvarint() (uint64, error)

	// ReadZInt32 Read a zig-zag-encoded variable-length integer.
	// See Also: DataOutput.writeZInt(int)
	ReadZInt32() (int64, error)

	// ReadUint64 Reads eight bytes and returns a long.
	// See Also: DataOutput.writeLong(long)
	ReadUint64() (uint64, error)

	// TODO: LUCENE-9047: Make the entire DataInput/DataOutput API little endian
	// Then this would just be `readLongs`?

	// ReadVInt64 Reads a long stored in variable-length format. Reads between one and nine bytes. Smaller values take fewer bytes. Negative numbers are not supported.
	// The format is described further in DataOutput.writeVInt(int).
	// See Also: DataOutput.writeVLong(long)
	//ReadVInt64() (uint64, error)

	// ReadZInt64 Read a zig-zag-encoded variable-length integer. Reads between one and ten bytes.
	// See Also: DataOutput.writeZLong(long)
	ReadZInt64() (int64, error)

	// ReadString Reads a string.
	// See Also: DataOutput.writeString(String)
	ReadString() (string, error)

	// ReadMapOfStrings Reads a Map<String,String> previously written with DataOutput.writeMapOfStrings(Map).
	// Returns: An immutable map containing the written contents.
	ReadMapOfStrings() (map[string]string, error)

	// ReadSetOfStrings Reads a Set<String> previously written with DataOutput.writeSetOfStrings(Set).
	// Returns: An immutable set containing the written contents.
	ReadSetOfStrings() (map[string]struct{}, error)

	SkipBytes(numBytes int) error
}
