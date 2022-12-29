package store

import (
	"encoding/binary"
	"fmt"

	"github.com/geange/lucene-go/core/util"
)

type DataInputDefaultConfig struct {
	ReadByte func() (byte, error)
	Read     func(p []byte) (n int, err error)
}

func NewDataInputDefault(cfg *DataInputDefaultConfig) *DataInputDefault {
	return &DataInputDefault{
		readByte: cfg.ReadByte,
		read:     cfg.Read,
		endian:   binary.BigEndian,
		buff:     make([]byte, 48),
	}
}

type DataInputDefault struct {
	readByte func() (byte, error)
	read     func(p []byte) (n int, err error)

	EOF    bool
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

func (d *DataInputDefault) Clone(cfg *DataInputDefaultConfig) *DataInputDefault {
	buff := make([]byte, len(d.buff))
	copy(buff, d.buff)

	skipBuffer := make([]byte, len(d.skipBuffer))
	copy(skipBuffer, d.skipBuffer)

	return &DataInputDefault{
		readByte:   cfg.ReadByte,
		read:       cfg.Read,
		EOF:        d.EOF,
		endian:     d.endian,
		buff:       buff,
		skipBuffer: skipBuffer,
	}
}

func (d *DataInputDefault) ReadByte() (byte, error) {
	if d.readByte != nil {
		return d.readByte()
	}
	bs := [1]byte{}
	_, err := d.read(bs[:])
	if err != nil {
		return 0, err
	}
	return bs[0], nil
}

func (d *DataInputDefault) ReadUint16() (uint16, error) {
	_, err := d.read(d.buff[:2])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint16(d.buff), nil
}

func (d *DataInputDefault) ReadUint32() (uint32, error) {
	_, err := d.read(d.buff[:4])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint32(d.buff), nil
}

func (d *DataInputDefault) ReadUvarint() (uint64, error) {
	num, err := binary.ReadUvarint(d)
	if err != nil {
		return 0, err
	}
	return num, err
}

func (d *DataInputDefault) ReadZInt32() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataInputDefault) ReadUint64() (uint64, error) {
	_, err := d.read(d.buff[:8])
	if err != nil {
		return 0, err
	}
	return d.endian.Uint64(d.buff), nil
}

func (d *DataInputDefault) ReadZInt64() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataInputDefault) ReadString() (string, error) {
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

	_, err = d.read(buf)
	if err != nil {
		return "", err
	}
	return string(d.buff[:length]), nil
}

func (d *DataInputDefault) ReadMapOfStrings() (map[string]string, error) {
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

func (d *DataInputDefault) ReadSetOfStrings() (map[string]struct{}, error) {
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
func (d *DataInputDefault) SkipBytes(numBytes int) error {
	if numBytes < 0 {
		return fmt.Errorf("numBytes must be >= 0, got %d", numBytes)
	}
	if len(d.skipBuffer) == 0 {
		d.skipBuffer = make([]byte, SKIP_BUFFER_SIZE)
	}
	for skipped := 0; skipped < numBytes; {
		step := util.Min(SKIP_BUFFER_SIZE, numBytes-skipped)
		if _, err := d.read(d.skipBuffer[0:step]); err != nil {
			return err
		}
		skipped += step
	}
	return nil
}

func (d *DataInputDefault) Close() error {
	return nil
}
