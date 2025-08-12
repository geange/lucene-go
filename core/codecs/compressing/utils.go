package compressing

import (
	"context"
	"errors"
	"math"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/zigzag"
)

// ReadZFloat
// Reads a float in a variable-length format. Reads between one and five bytes.
// Small integral values typically take fewer bytes.
func ReadZFloat(ctx context.Context, in store.DataInput) (float32, error) {
	b, err := in.ReadByte()
	if err != nil {
		return 0, err
	}
	if b == 0xFF {
		// negative value
		num, err := in.ReadUint32(ctx)
		if err != nil {
			return 0, err
		}

		return math.Float32frombits(num), nil
	}

	// small integer [-1..125]
	if b&0x80 != 0 {
		return float32((b & 0x7f) - 1), nil
	}

	// positive float
	b1, err := in.ReadUint16(ctx)
	if err != nil {
		return 0, err
	}
	b2, err := in.ReadByte()
	if err != nil {
		return 0, err
	}
	bits := uint32(b)<<24 | uint32(b1&0xFFFF)<<8 | uint32(b2)
	return math.Float32frombits(bits), nil
}

// ReadZDouble
// Reads a double in a variable-length format. Reads between one and nine bytes.
// Small integral values typically take fewer bytes.
func ReadZDouble(ctx context.Context, in store.DataInput) (float64, error) {
	b, err := in.ReadByte()
	if err != nil {
		return 0, err
	}

	if b == 0xFF {
		// negative value
		// negative value
		num, err := in.ReadUint64(ctx)
		if err != nil {
			return 0, err
		}

		return math.Float64frombits(num), nil
	}

	if b == 0xFE {
		num, err := in.ReadUint32(ctx)
		if err != nil {
			return 0, err
		}

		return float64(math.Float32frombits(num)), nil
	}

	if b&0x80 != 0 {
		// small integer [-1..124]
		return float64(b&0x7F) - 1, nil
	}

	// positive double
	b1, err := in.ReadUint32(ctx)
	if err != nil {
		return 0, err
	}
	b2, err := in.ReadUint16(ctx)
	if err != nil {
		return 0, err
	}
	b3, err := in.ReadByte()
	if err != nil {
		return 0, err
	}
	bits := uint64(b)<<56 | uint64(b1)<<24 | uint64(b2)<<8 | uint64(b3)
	return math.Float64frombits(bits), nil
}

// ReadTLong
// Reads a long in a variable-length format. Reads between one andCorePropLo nine bytes.
// Small values typically take fewer bytes.
func ReadTLong(ctx context.Context, in store.DataInput) (int64, error) {
	header, err := in.ReadByte()
	if err != nil {
		return 0, err
	}
	bits := uint64(header & 0x1F)
	if (header & 0x20) != 0 {
		// continuation bit
		n, err := in.ReadUvarint(ctx)
		if err != nil {
			return 0, err
		}
		bits |= n << 5
	}
	l := zigzag.Decode(bits)

	switch header & byte(DAY_ENCODING) {
	case SECOND_ENCODING:
		l *= int64(SECOND)
	case HOUR_ENCODING:
		l *= int64(HOUR)
	case DAY_ENCODING:
		l *= int64(DAY)
	case 0:
		// uncompressed
	default:
		return 0, errors.New("assert flag")
	}
	return l, nil
}

// WriteZFloat
// Writes a float in a variable-length format. Writes between one and five bytes. Small integral values
// typically take fewer bytes.
// ZFloat --> Header, Bytes*?
// Header --> Uint8. When it is equal to 0xFF then the value is negative and stored in the next 4 bytes.
//
//	Otherwise if the first bit is set then the other bits in the header encode the value plus
//	one and no other bytes are read. Otherwise, the value is a positive float value whose first
//	byte is the header, and 3 bytes need to be read to complete it.
//
// Bytes --> Potential additional bytes to read depending on the header.
func WriteZFloat(ctx context.Context, out store.DataOutput, f float32) error {
	intVal := int(f)
	floatBits := math.Float32bits(f)

	if f == float32(intVal) && intVal >= -1 && intVal <= 0x7D && floatBits != NEGATIVE_ZERO_FLOAT {
		// small integer value [-1..125]: single byte
		if err := out.WriteByte(byte(0x80 | (1 + intVal))); err != nil {
			return err
		}
	} else if (floatBits >> 31) == 0 {
		// other positive floats: 4 bytes
		if err := out.WriteUint32(ctx, floatBits); err != nil {
			return err
		}
	} else {
		// other negative float: 5 bytes
		if err := out.WriteByte(0xFF); err != nil {
			return err
		}
		if err := out.WriteUint32(ctx, floatBits); err != nil {
			return err
		}
	}
	return nil
}

// WriteZDouble
// Writes a float in a variable-length format. Writes between one and five bytes. Small integral values typically take fewer bytes.
// ZFloat --> Header, Bytes*?
// Header --> Uint8. When it is equal to 0xFF then the value is negative and stored in the next 8 bytes. When it is equal to 0xFE then the value is stored as a float in the next 4 bytes. Otherwise if the first bit is set then the other bits in the header encode the value plus one and no other bytes are read. Otherwise, the value is a positive float value whose first byte is the header, and 7 bytes need to be read to complete it.
// Bytes --> Potential additional bytes to read depending on the header.
func WriteZDouble(ctx context.Context, out store.DataOutput, d float64) error {
	intVal := int(d)
	doubleBits := math.Float64bits(d)

	if d == float64(intVal) && intVal >= -1 && intVal <= 0x7C && doubleBits != NEGATIVE_ZERO_DOUBLE {
		// small integer value [-1..124]: single byte
		if err := out.WriteByte(byte(0x80 | (intVal + 1))); err != nil {
			return err
		}
		return nil
	} else if d == float64(float32(d)) {
		// d has an accurate float representation: 5 bytes
		if err := out.WriteByte(0xFE); err != nil {
			return err
		}

		if err := out.WriteUint32(ctx, math.Float32bits(float32(d))); err != nil {
			return err
		}
	} else if (doubleBits >> 63) == 0 {
		// other positive doubles: 8 bytes
		if err := out.WriteUint64(ctx, doubleBits); err != nil {
			return err
		}
	} else {
		// other negative doubles: 9 bytes
		if err := out.WriteByte(0xFF); err != nil {
			return err
		}
		if err := out.WriteUint64(ctx, doubleBits); err != nil {
			return err
		}
	}
	return nil
}

// WriteTLong
// Writes a long in a variable-length format. Writes between one and ten bytes. Small values or values representing timestamps with day, hour or second precision typically require fewer bytes.
// ZLong --> Header, Bytes*?
// Header --> The first two bits indicate the compression scheme:
// 00 - uncompressed
// 01 - multiple of 1000 (second)
// 10 - multiple of 3600000 (hour)
// 11 - multiple of 86400000 (day)
// Then the next bit is a continuation bit, indicating whether more bytes need to be read, and the last 5 bits are the lower bits of the encoded value. In order to reconstruct the value, you need to combine the 5 lower bits of the header with a vLong in the next bytes (if the continuation bit is set to 1). Then zigzag-decode it and finally multiply by the multiple corresponding to the compression scheme.
// Bytes --> Potential additional bytes to read depending on the header.
// T for "timestamp"
func WriteTLong(ctx context.Context, out store.DataOutput, l int64) error {
	var header byte
	if l%SECOND != 0 {
		header = 0
	} else if l%DAY == 0 {
		// timestamp with day precision
		header = DAY_ENCODING
		l /= DAY
	} else if l%HOUR == 0 {
		// timestamp with hour precision, or day precision with a timezone
		header = HOUR_ENCODING
		l /= HOUR
	} else {
		// timestamp with second precision
		header = SECOND_ENCODING
		l /= SECOND
	}

	zigZagL := zigzag.Encode(l)
	header = header | (byte(zigZagL) & 0x1F) // last 5 bits
	upperBits := zigZagL >> 5
	if upperBits != 0 {
		header |= 0x20
	}
	if err := out.WriteByte(header); err != nil {
		return err
	}
	if upperBits != 0 {
		if err := out.WriteUvarint(ctx, upperBits); err != nil {
			return err
		}
	}
	return nil
}
