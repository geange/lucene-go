package store

import (
	"encoding/binary"
	"fmt"
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

func NewReaderX(reader io.Reader) *ReaderX {
	return &ReaderX{
		reader: reader,
		endian: binary.BigEndian,
		buff:   make([]byte, 48),
	}
}

type ReaderX struct {
	reader io.Reader
	endian binary.ByteOrder
	buff   []byte

	// This buffer is used to skip over bytes with the default implementation of
	// skipBytes. The reason why we reader to use an instance member instead of
	// sharing a single instance across threads is that some delegating
	// implementations of DataInput might want to reuse the provided buffer in
	// order to eg. update the checksum. If we shared the same buffer across
	// threads, then another thread might update the buffer while the checksum is
	// being computed, making it invalid. See LUCENE-5583 for more information.
	skipBuffer []byte
}

func (d *ReaderX) Clone(reader io.Reader) *ReaderX {
	buff := make([]byte, len(d.buff))
	copy(buff, d.buff)

	skipBuffer := make([]byte, len(d.skipBuffer))
	copy(skipBuffer, d.skipBuffer)

	return &ReaderX{
		reader:     reader,
		endian:     d.endian,
		buff:       buff,
		skipBuffer: skipBuffer,
	}
}

func (d *ReaderX) ReadByte() (byte, error) {
	_, err := d.reader.Read(d.buff[:1])
	if err != nil {
		return 0, err
	}
	return d.buff[0], nil
}

func (d *ReaderX) ReadUint16() (uint16, error) {
	_, err := d.reader.Read(d.buff[:2])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint16(d.buff), nil
}

func (d *ReaderX) ReadUint32() (uint32, error) {
	_, err := d.reader.Read(d.buff[:4])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint32(d.buff), nil
}

func (d *ReaderX) ReadUvarint() (uint64, error) {
	num, err := binary.ReadUvarint(d)
	if err != nil {
		return 0, err
	}
	return num, err
}

func (d *ReaderX) ReadZInt32() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ReaderX) ReadUint64() (uint64, error) {
	_, err := d.reader.Read(d.buff[:8])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint64(d.buff), nil
}

func (d *ReaderX) ReadZInt64() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ReaderX) ReadString() (string, error) {
	num, err := d.ReadUvarint()
	if err != nil {
		return "", err
	}

	length := int(num)

	var buf []byte
	if len(d.buff) > length {
		buf = d.buff[:length]
	} else {
		buf = make([]byte, length)
	}

	_, err = d.reader.Read(buf)
	if err != nil {
		return "", err
	}
	return string(d.buff[:length]), nil
}

func (d *ReaderX) ReadMapOfStrings() (map[string]string, error) {
	count, err := d.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return map[string]string{}, nil
	}

	values := make(map[string]string, int(count))

	for i := 0; i < int(count); i++ {
		k, err := d.ReadString()
		if err != nil {
			return nil, err
		}

		v, err := d.ReadString()
		if err != nil {
			return nil, err
		}

		values[k] = v
	}
	return values, nil
}

func (d *ReaderX) ReadSetOfStrings() (map[string]struct{}, error) {
	count, err := d.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return map[string]struct{}{}, nil
	}

	values := make(map[string]struct{}, int(count))

	for i := 0; i < int(count); i++ {
		k, err := d.ReadString()
		if err != nil {
			return nil, err
		}
		values[k] = struct{}{}
	}
	return values, nil
}

// SkipBytes Closer Skip over numBytes bytes. The contract on this method is that it should have the
// same behavior as reading the same number of bytes into a buffer and discarding its content.
// Negative values of numBytes are not supported.
func (d *ReaderX) SkipBytes(numBytes int) error {
	if numBytes < 0 {
		return fmt.Errorf("numBytes must be >= 0, got %d", numBytes)
	}
	if len(d.skipBuffer) == 0 {
		d.skipBuffer = make([]byte, SKIP_BUFFER_SIZE)
	}
	for skipped := 0; skipped < numBytes; {
		step := min(SKIP_BUFFER_SIZE, numBytes-skipped)
		if _, err := d.reader.Read(d.skipBuffer[0:step]); err != nil {
			return err
		}
		skipped += step
	}
	return nil
}

func (d *ReaderX) Close() error {
	return nil
}

// DataOutput Abstract base class for performing write operations of Lucene's low-level data types.
// DataOutput may only be used from one thread, because it is not thread safe (it keeps internal state like file pos).
type DataOutput interface {
	// ByteWriter Write a single byte.
	// The most primitive data type is an eight-bit byte. Files are accessed as sequences of bytes.
	// All other data types are defined as sequences of bytes, so file formats are byte-order independent.
	// See Also: IndexInput.readByte()
	io.ByteWriter

	// Writer Writes an array of bytes.
	io.Writer

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

type WriterX struct {
	writer     io.Writer
	endian     binary.ByteOrder
	buffer     []byte
	copyBuffer []byte
}

func NewWriterX(writer io.Writer) *WriterX {
	return &WriterX{
		writer: writer,
		endian: binary.BigEndian,
		buffer: make([]byte, 48),
	}
}

func (d *WriterX) WriteByte(c byte) error {
	//if w, ok := d.writer.(io.ByteWriter); ok {
	//	return w.WriteByte(c)
	//}
	_, err := d.writer.Write([]byte{c})
	return err
}

func (d *WriterX) WriteUint32(i uint32) error {
	d.endian.PutUint32(d.buffer, i)
	_, err := d.writer.Write(d.buffer[:4])
	return err
}

func (d *WriterX) WriteUint16(i uint16) error {
	d.endian.PutUint16(d.buffer, i)
	_, err := d.writer.Write(d.buffer[:2])
	return err
}

func (d *WriterX) WriteUvarint(i uint64) error {
	num := binary.PutUvarint(d.buffer, i)
	_, err := d.writer.Write(d.buffer[:num])
	return err
}

func (d *WriterX) WriteZInt32(i int32) error {
	//TODO implement me
	panic("implement me")
}

func (d *WriterX) WriteUint64(i uint64) error {
	d.endian.PutUint64(d.buffer, i)
	_, err := d.writer.Write(d.buffer[:8])
	return err
}

func (d *WriterX) WriteZInt64(i int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *WriterX) WriteString(s string) error {
	err := d.WriteUvarint(uint64(len([]rune(s))))
	if err != nil {
		return err
	}
	_, err = d.writer.Write([]byte(s))
	return err
}

const (
	COPY_BUFFER_SIZE = 16384
)

func (d *WriterX) CopyBytes(input DataInput, numBytes int) error {
	left := numBytes
	if len(d.copyBuffer) == 0 {
		d.copyBuffer = make([]byte, COPY_BUFFER_SIZE)
	}

	for left > 0 {
		var toCopy int
		if left > COPY_BUFFER_SIZE {
			toCopy = COPY_BUFFER_SIZE
		} else {
			toCopy = left
		}
		_, err := input.Read(d.copyBuffer[:toCopy])
		if err != nil {
			return err
		}
		_, err = d.writer.Write(d.copyBuffer[:toCopy])
		if err != nil {
			return err
		}
		left -= toCopy
	}
	return nil
}

func (d *WriterX) WriteMapOfStrings(values map[string]string) error {
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

func (d *WriterX) WriteSetOfStrings(values map[string]struct{}) error {
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
