package store

import (
	"encoding/binary"
	"io"
)

// DataInput Abstract base class for performing read operations of Lucene's low-level data types.
// DataInput may only be used from one thread, because it is not thread safe (it keeps internal state
// like file position). To allow multithreaded use, every DataInput instance must be cloned before used
// in another thread. Subclasses must therefore implement clone(), returning a new DataInput which operates
// on the same underlying resource, but positioned independently.
type DataInput interface {
	// ReadByte Reads and returns a single byte.
	// See Also: DataOutput.writeByte(byte)
	ReadByte() (byte, error)

	// ReadBytes Reads a specified number of bytes into an array.
	ReadBytes(b []byte) error

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

	// SkipBytes Closer Skip over numBytes bytes. The contract on this method is that it should have the
	// same behavior as reading the same number of bytes into a buffer and discarding its content.
	// Negative values of numBytes are not supported.
	SkipBytes(numBytes int) error
}

type DataInputNeed interface {
	ReadByte() (byte, error)
	ReadBytes(b []byte) error
}

var _ DataInput = &DataInputImp{}
var _ io.ByteReader = &DataInputImp{}

type DataInputImp struct {
	DataInputNeed

	EOF    bool
	endian binary.ByteOrder
	buff   []byte

	// This buffer is used to skip over bytes with the default implementation of
	// skipBytes. The reason why we need to use an instance member instead of
	// sharing a single instance across threads is that some delegating
	// implementations of DataInput might want to reuse the provided buffer in
	// order to eg. update the checksum. If we shared the same buffer across
	// threads, then another thread might update the buffer while the checksum is
	// being computed, making it invalid. See LUCENE-5583 for more information.
	skipBuffer []byte
}

func NewDataInputImp(need DataInputNeed) *DataInputImp {
	return &DataInputImp{
		DataInputNeed: need,
		endian:        binary.BigEndian,
		buff:          make([]byte, 48),
	}
}

func (d *DataInputImp) ReadUint16() (uint16, error) {
	err := d.ReadBytes(d.buff[:2])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint16(d.buff), nil
}

func (d *DataInputImp) ReadUint32() (uint32, error) {
	err := d.ReadBytes(d.buff[:4])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint32(d.buff), nil
}

func (d *DataInputImp) ReadUvarint() (uint64, error) {
	num, err := binary.ReadUvarint(d)
	if err != nil {
		return 0, err
	}
	return num, err
}

func (d *DataInputImp) ReadZInt32() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataInputImp) ReadUint64() (uint64, error) {
	err := d.ReadBytes(d.buff[:8])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint64(d.buff), nil
}

func (d *DataInputImp) ReadZInt64() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataInputImp) ReadString() (string, error) {
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

	err = d.ReadBytes(buf)
	if err != nil {
		return "", err
	}
	return string(d.buff[:length]), nil
}

func (d *DataInputImp) ReadMapOfStrings() (map[string]string, error) {
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

func (d *DataInputImp) ReadSetOfStrings() (map[string]struct{}, error) {
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

func (d *DataInputImp) SkipBytes(numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataInputImp) Close() error {
	//TODO implement me
	panic("implement me")
}
