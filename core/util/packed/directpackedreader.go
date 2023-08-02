package packed

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"io"
	"math"
)

var _ Reader = &DirectPackedReader{}

// DirectPackedReader
// Reads directly from disk on each get
// just for back compat, use DirectReader/DirectWriter for more efficient impl
type DirectPackedReader struct {
	in           store.IndexInput
	bitsPerValue int
	startPointer int64
	valueMask    uint64
	valueCount   int
}

func (d *DirectPackedReader) GetBulk(index int, arr []uint64) int {
	for i := range arr {
		n, err := d.Get(index + i)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return i
			}
			return 0
		}
		arr[i] = n
	}
	return len(arr)
}

func (d *DirectPackedReader) Size() int {
	return d.valueCount
}

func NewDirectPackedReader(bitsPerValue, valueCount int, in store.IndexInput) *DirectPackedReader {
	reader := &DirectPackedReader{
		in:           in,
		bitsPerValue: bitsPerValue,
		startPointer: in.GetFilePointer(),
		valueCount:   valueCount,
	}

	if bitsPerValue == 64 {
		reader.valueMask = math.MaxUint64
	} else {
		reader.valueMask = (1 << bitsPerValue) - 1
	}
	return reader
}

func (d *DirectPackedReader) Get(index int) (uint64, error) {
	majorBitPos := index * d.bitsPerValue
	elementPos := int64(majorBitPos >> 3)

	if _, err := d.in.Seek(d.startPointer+elementPos, io.SeekStart); err != nil {
		return 0, err
	}

	bitPos := majorBitPos & 7
	// round up bits to a multiple of 8 to find total bytes needed to read
	// 将位数四舍五入到8的倍数，以找到读取所需的总字节数
	// roundedBits = (bitPos+d.bitsPerValue+7) / 8
	roundedBits := int(uint64(bitPos+d.bitsPerValue+7) & (^uint64(7)))
	// the number of extra bits read at the end to shift out
	shiftRightBits := roundedBits - bitPos - d.bitsPerValue

	var rawValue uint64

	ctx := context.TODO()

	switch roundedBits >> 3 {
	case 1:
		v, err := d.in.ReadByte()
		if err != nil {
			return 0, err
		}
		rawValue = uint64(v)

	case 2:
		v, err := d.in.ReadUint16(ctx)
		if err != nil {
			return 0, err
		}
		rawValue = uint64(v)

	case 3:
		n1, err := d.in.ReadUint16(ctx)
		if err != nil {
			return 0, err
		}
		n2, err := d.in.ReadByte()
		if err != nil {
			return 0, err
		}
		rawValue = uint64(n1)<<8 | uint64(n2)

	case 4:
		v, err := d.in.ReadUint32(ctx)
		if err != nil {
			return 0, err
		}
		rawValue = uint64(v)

	case 5:
		n1, err := d.in.ReadUint32(ctx)
		if err != nil {
			return 0, err
		}
		n2, err := d.in.ReadByte()
		if err != nil {
			return 0, err
		}
		rawValue = uint64(n1)<<8 | uint64(n2)

	case 6:
		n1, err := d.in.ReadUint32(ctx)
		if err != nil {
			return 0, err
		}
		n2, err := d.in.ReadUint16(ctx)
		if err != nil {
			return 0, err
		}
		rawValue = uint64(n1)<<16 | uint64(n2)

	case 7:
		n1, err := d.in.ReadUint32(ctx)
		if err != nil {
			return 0, err
		}
		n2, err := d.in.ReadUint16(ctx)
		if err != nil {
			return 0, err
		}
		n3, err := d.in.ReadByte()
		if err != nil {
			return 0, err
		}
		rawValue = uint64(n1)<<24 | uint64(n2)<<8 | uint64(n3)

	case 8:
		v, err := d.in.ReadUint64(ctx)
		if err != nil {
			return 0, err
		}
		rawValue = v

	case 9:
		// We must be very careful not to shift out relevant bits. So we account for right shift
		// we would normally do on return here, and reset it.
		n1, err := d.in.ReadUint64(ctx)
		if err != nil {
			return 0, err
		}
		n2, err := d.in.ReadByte()
		if err != nil {
			return 0, err
		}
		rawValue = n1<<(8-shiftRightBits) | uint64(n2)>>shiftRightBits
		shiftRightBits = 0

	default:
		return 0, errors.New("bitsPerValue too large")
	}

	return (rawValue >> shiftRightBits) & d.valueMask, nil
}
