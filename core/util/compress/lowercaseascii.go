package compress

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/store"
)

var (
	LowercaseAsciiCompression = &LowercaseAscii{}
)

type LowercaseAscii struct {
}

func (*LowercaseAscii) Compress(ctx context.Context, in, tmp []byte, out store.DataOutput) (bool, error) {
	size := len(in)

	if size < 8 {
		return false, nil
	}

	// 1. Count exceptions and fail compression if there are too many of them.
	maxExceptions := size >> 5
	previousExceptionIndex := 0
	numExceptions := 0
	for i := 0; i < size; i++ {
		b := in[i]
		if isCompressible(b) == false {
			for i-previousExceptionIndex > 0xFF {
				numExceptions++
				previousExceptionIndex += 0xFF
			}
			numExceptions++
			if numExceptions > maxExceptions {
				return false, nil
			}
			previousExceptionIndex = i
		}
	}

	// 2. Now move all bytes to the [0,0x40) range (6 bits). This loop gets auto-vectorized on JDK13+.
	compressedLen := size - (size >> 2) // ignores exceptions
	for i := 0; i < size; i++ {
		b := (in[i]) + 1
		tmp[i] = (b & 0x1F) | ((b & 0x40) >> 1)
	}

	// 3. Now pack the bytes so that we record 4 ASCII chars in 3 bytes
	o := 0
	for i := compressedLen; i < size; i++ {
		tmp[o] |= (tmp[i] & 0x30) << 2 // bits 4-5
		o++
	}
	for i := compressedLen; i < size; i++ {
		tmp[o] |= (tmp[i] & 0x0C) << 4 // bits 2-3
		o++
	}
	for i := compressedLen; i < size; i++ {
		tmp[o] |= (tmp[i] & 0x03) << 6 // bits 0-1
		o++
	}

	if _, err := out.Write(tmp[:compressedLen]); err != nil {
		return false, err
	}

	// 4. Finally record exceptions
	if err := out.WriteUvarint(ctx, uint64(numExceptions)); err != nil {
		return false, err
	}
	if numExceptions > 0 {
		previousExceptionIndex = 0
		numExceptions2 := 0
		for i := 0; i < size; i++ {
			b := in[i]
			if isCompressible(b) == false {
				for i-previousExceptionIndex > 0xFF {
					// We record deltas between exceptions as bytes, so we need to create
					// "artificial" exceptions if the delta between two of them is greater
					// than the maximum unsigned byte value.
					if err := out.WriteByte(0xFF); err != nil {
						return false, err
					}
					previousExceptionIndex += 0xFF
					if err := out.WriteByte(in[previousExceptionIndex]); err != nil {
						return false, err
					}
					numExceptions2++
				}
				if err := out.WriteByte(byte(i - previousExceptionIndex)); err != nil {
					return false, err
				}
				previousExceptionIndex = i
				if err := out.WriteByte(b); err != nil {
					return false, err
				}
				numExceptions2++
			}
		}
		if numExceptions != numExceptions2 {
			return false, errors.New("illegal state exception")
		}
	}
	return true, nil
}

func isCompressible(b byte) bool {
	high3Bits := (b + 1) &^ 0x1F
	return high3Bits == 0x20 || high3Bits == 0x60
}

func (*LowercaseAscii) Decompress(ctx context.Context, in store.DataInput, out []byte) error {
	size := len(out)
	saved := size >> 2
	compressedLen := size - saved

	// 1. Copy the packed bytes
	if _, err := in.Read(out[:compressedLen]); err != nil {
		return err
	}

	// 2. Restore the leading 2 bits of each packed byte into whole bytes
	for i := 0; i < saved; i++ {
		out[compressedLen+i] = ((out[i] & 0xC0) >> 2) |
			((out[saved+i] & 0xC0) >> 4) |
			((out[(saved<<1)+i] & 0xC0) >> 6)
	}

	// 3. Move back to the original range. This loop gets auto-vectorized on JDK13+.
	for i := 0; i < size; i++ {
		b := out[i]
		out[i] = ((b & 0x1F) | 0x20 | ((b & 0x20) << 1)) - 1
	}

	// 4. Restore exceptions
	numExceptions, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	i := 0
	for exception := 0; exception < int(numExceptions); exception++ {
		b, err := in.ReadByte()
		if err != nil {
			return err
		}
		i += int(b)

		o, err := in.ReadByte()
		if err != nil {
			return err
		}
		out[i] = o
	}
	return nil
}
