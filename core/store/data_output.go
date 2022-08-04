package store

import "encoding/binary"

// DataOutput Abstract base class for performing write operations of Lucene's low-level data types.
// DataOutput may only be used from one thread, because it is not thread safe (it keeps internal state like file position).
type DataOutput interface {

	// WriteByte Writes a single byte.
	// The most primitive data type is an eight-bit byte. Files are accessed as sequences of bytes.
	// All other data types are defined as sequences of bytes, so file formats are byte-order independent.
	// See Also: IndexInput.readByte()
	WriteByte(b byte) error

	// WriteBytes Writes an array of bytes.
	WriteBytes(b []byte) error

	// WriteUint32 Writes an int as four bytes.
	// 32-bit unsigned integer written as four bytes, high-order bytes first.
	// See Also: DataInput.readInt()
	WriteUint32(i uint32) error

	// WriteUint16 Writes a short as two bytes.
	// See Also: DataInput.readShort()
	WriteUint16(i uint16) error

	// WriteUvarint Writes an int in a variable-length format. Writes between one and five bytes. Smaller
	// values take fewer bytes. Negative numbers are supported, but should be avoided.
	// VByte is a variable-length format for positive integers is defined where the high-order bit of each
	// byte indicates whether more bytes remain to be read. The low-order seven bits are appended as
	// increasingly more significant bits in the resulting integer value. Thus values from zero to 127 may
	// be stored in a single byte, values from 128 to 16,383 may be stored in two bytes, and so on.
	WriteUvarint(i uint64) error

	// WriteZInt32 Write a zig-zag-encoded variable-length integer. This is typically useful to write small
	// signed ints and is equivalent to calling writeVInt(BitUtil.zigZagEncode(i)).
	// See Also: DataInput.readZInt()
	WriteZInt32(i int32) error

	// WriteUint64 Writes a long as eight bytes.
	// 64-bit unsigned integer written as eight bytes, high-order bytes first.
	// See Also: DataInput.readLong()
	WriteUint64(i uint64) error

	// WriteVInt64 Writes an long in a variable-length format. Writes between one and nine bytes.
	// Smaller values take fewer bytes. Negative numbers are not supported.
	// The format is described further in writeVInt(int).
	// See Also: DataInput.readVLong()
	//WriteVInt64(i uint64) error

	// WriteZInt64 Write a zig-zag-encoded variable-length long. Writes between one and ten bytes. This is typically
	// useful to write small signed ints.
	// See Also: DataInput.readZLong()
	WriteZInt64(i int64) error

	// WriteString Writes a string.
	// Writes strings as UTF-8 encoded bytes. First the length, in bytes, is written as a VInt, followed by the bytes.
	// See Also: DataInput.readString()
	WriteString(s string) error

	// CopyBytes Copy numBytes bytes from input to ourself.
	CopyBytes(input DataInput, numBytes int) error

	// WriteMapOfStrings Writes a String map.
	// First the size is written as an vInt, followed by each key-value pair written as two consecutive Strings.
	WriteMapOfStrings(values map[string]string) error

	// WriteSetOfStrings Writes a String set.
	//First the size is written as an vInt, followed by each value written as a String.
	WriteSetOfStrings(values map[string]struct{}) error
}

type DataOutputNeed interface {
	// WriteBytes Writes an array of bytes.
	WriteBytes(b []byte) error
}

var _ DataOutput = &DataOutputImp{}

type DataOutputImp struct {
	DataOutputNeed

	endian binary.ByteOrder
	buffer []byte
}

func NewDataOutputImp(need DataOutputNeed) *DataOutputImp {
	return &DataOutputImp{
		DataOutputNeed: need,
		endian:         binary.BigEndian,
		buffer:         make([]byte, 48),
	}
}

func (d *DataOutputImp) WriteByte(b byte) error {
	return d.WriteBytes([]byte{b})
}

func (d *DataOutputImp) WriteUint32(i uint32) error {
	d.endian.PutUint32(d.buffer, i)
	return d.WriteBytes(d.buffer[:4])
}

func (d *DataOutputImp) WriteUint16(i uint16) error {
	d.endian.PutUint16(d.buffer, i)
	return d.WriteBytes(d.buffer[:2])
}

func (d *DataOutputImp) WriteUvarint(i uint64) error {
	num := binary.PutUvarint(d.buffer, i)
	return d.WriteBytes(d.buffer[:num])
}

func (d *DataOutputImp) WriteZInt32(i int32) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataOutputImp) WriteUint64(i uint64) error {
	d.endian.PutUint64(d.buffer, i)
	return d.WriteBytes(d.buffer[:8])
}

func (d *DataOutputImp) WriteZInt64(i int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataOutputImp) WriteString(s string) error {
	err := d.WriteUvarint(uint64(len([]rune(s))))
	if err != nil {
		return err
	}
	return d.WriteBytes([]byte(s))
}

func (d *DataOutputImp) CopyBytes(input DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataOutputImp) WriteMapOfStrings(values map[string]string) error {
	if err := d.WriteUvarint(uint64(len(values))); err != nil {
		return err
	}

	for k, v := range values {
		if err := d.WriteString(k); err != nil {
			return err
		}
		if err := d.WriteString(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DataOutputImp) WriteSetOfStrings(values map[string]struct{}) error {
	if err := d.WriteUvarint(uint64(len(values))); err != nil {
		return err
	}

	for k := range values {
		if err := d.WriteString(k); err != nil {
			return err
		}
	}
	return nil
}
