package store

import (
	"encoding/binary"
)

type DataOutputDefaultConfig struct {
	WriteByte  func(b byte) error
	WriteBytes func(b []byte) (int, error)
}

type DataOutputDefault struct {
	// writeByte Writes a single byte.
	// The most primitive data type is an eight-bit byte. Files are accessed as sequences of bytes.
	// All other data types are defined as sequences of bytes, so file formats are byte-order independent.
	// See Also: IndexInput.readByte()
	writeByte func(b byte) error

	// writeBytes Writes an array of bytes.
	writeBytes func(b []byte) (int, error)

	endian     binary.ByteOrder
	buffer     []byte
	copyBuffer []byte
}

func NewDataOutputDefault(cfg *DataOutputDefaultConfig) *DataOutputDefault {
	return &DataOutputDefault{
		writeByte:  cfg.WriteByte,
		writeBytes: cfg.WriteBytes,
		endian:     binary.BigEndian,
		buffer:     make([]byte, 48),
	}
}

func (d *DataOutputDefault) WriteByte(b byte) error {
	if d.writeByte == nil {
		_, err := d.writeBytes([]byte{b})
		return err
	}
	return d.writeByte(b)
}

func (d *DataOutputDefault) WriteUint32(i uint32) error {
	d.endian.PutUint32(d.buffer, i)
	_, err := d.writeBytes(d.buffer[:4])
	return err
}

func (d *DataOutputDefault) WriteUint16(i uint16) error {
	d.endian.PutUint16(d.buffer, i)
	_, err := d.writeBytes(d.buffer[:2])
	return err
}

func (d *DataOutputDefault) WriteUvarint(i uint64) error {
	num := binary.PutUvarint(d.buffer, i)
	_, err := d.writeBytes(d.buffer[:num])
	return err
}

func (d *DataOutputDefault) WriteZInt32(i int32) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataOutputDefault) WriteUint64(i uint64) error {
	d.endian.PutUint64(d.buffer, i)
	_, err := d.writeBytes(d.buffer[:8])
	return err
}

func (d *DataOutputDefault) WriteZInt64(i int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataOutputDefault) WriteString(s string) error {
	err := d.WriteUvarint(uint64(len([]rune(s))))
	if err != nil {
		return err
	}
	_, err = d.writeBytes([]byte(s))
	return err
}

const (
	COPY_BUFFER_SIZE = 16384
)

func (d *DataOutputDefault) CopyBytes(input DataInput, numBytes int) error {
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
		_, err = d.writeBytes(d.copyBuffer[:toCopy])
		if err != nil {
			return err
		}
		left -= toCopy
	}
	return nil
}

func (d *DataOutputDefault) WriteMapOfStrings(values map[string]string) error {
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

func (d *DataOutputDefault) WriteSetOfStrings(values map[string]struct{}) error {
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
