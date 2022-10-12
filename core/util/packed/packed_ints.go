package packed

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
	"github.com/pkg/errors"
	"math"
	"math/bits"
)

const (
	// FASTEST At most 700% memory overhead, always select a direct implementation.
	FASTEST = 7.0

	// FAST At most 50% memory overhead, always select a reasonably fast implementation.
	FAST = 0.5

	// DEFAULT At most 25% memory overhead.
	DEFAULT = 0.25

	// COMPACT No memory overhead at all, but the returned implementation may be slow.
	COMPACT = 0.0

	// DEFAULT_BUFFER_SIZE Default amount of memory to use for bulk operations.
	DEFAULT_BUFFER_SIZE = 1024

	CODEC_NAME = "PackedInts"

	VERSION_MONOTONIC_WITHOUT_ZIGZAG = 2
	VERSION_START                    = VERSION_MONOTONIC_WITHOUT_ZIGZAG
	VERSION_CURRENT                  = VERSION_MONOTONIC_WITHOUT_ZIGZAG
)

// PackedInts Simplistic compression for array of unsigned long values.
// Each value is >= 0 and <= a specified maximum value. The values are stored as packed ints,
// with each value consuming a fixed number of bits.
// lucene.internal
type PackedInts struct {
}

// FormatAndBits Simple class that holds a format and a number of bits per value.
type FormatAndBits struct {
	format       Format
	bitsPerValue int
}

func NewFormatAndBits(format Format, bitsPerValue int) *FormatAndBits {
	return &FormatAndBits{format: format, bitsPerValue: bitsPerValue}
}

// Try to find the PackedInts.Format and number of bits per value that would restore from disk the
// fastest reader whose overhead is less than acceptableOverheadRatio.
// The acceptableOverheadRatio parameter makes sense for random-access PackedInts.Readers.
// In case you only plan to perform sequential access on this stream later on, you should probably use COMPACT.
// If you don't know how many values you are going to write, use valueCount = -1.
func fastestFormatAndBits(valueCount, bitsPerValue int, acceptableOverheadRatio float64) *FormatAndBits {
	if valueCount == -1 {
		valueCount = math.MaxInt32
	}

	acceptableOverheadRatio = Max(COMPACT, acceptableOverheadRatio)
	acceptableOverheadRatio = Min(FASTEST, acceptableOverheadRatio)
	acceptableOverheadPerValue := acceptableOverheadRatio * float64(bitsPerValue) // in bits

	maxBitsPerValue := bitsPerValue + int(acceptableOverheadPerValue)

	actualBitsPerValue := -1

	var format Format
	format = FormatPacked

	if bitsPerValue <= 8 && maxBitsPerValue >= 8 {
		actualBitsPerValue = 8
	} else if bitsPerValue <= 16 && maxBitsPerValue >= 16 {
		actualBitsPerValue = 16
	} else if bitsPerValue <= 32 && maxBitsPerValue >= 32 {
		actualBitsPerValue = 32
	} else if bitsPerValue <= 64 && maxBitsPerValue >= 64 {
		actualBitsPerValue = 64
	} else if valueCount <= Packed8ThreeBlocksMaxSize && bitsPerValue <= 24 && maxBitsPerValue >= 24 {
		actualBitsPerValue = 24
	} else if valueCount <= Packed8ThreeBlocksMaxSize && bitsPerValue <= 48 && maxBitsPerValue >= 48 {
		actualBitsPerValue = 48
	} else {
		for bpv := bitsPerValue; bpv <= maxBitsPerValue; bpv++ {
			if FormatPackedSingleBlock.IsSupported(bpv) {
				overhead := FormatPackedSingleBlock.OverheadPerValue(bpv)
				acceptableOverhead := acceptableOverheadPerValue + float64(bitsPerValue) - float64(bpv)
				if overhead <= acceptableOverhead {
					actualBitsPerValue = bpv
					format = FormatPackedSingleBlock
					break
				}
			}
		}
		if actualBitsPerValue < 0 {
			actualBitsPerValue = bitsPerValue
		}
	}

	return NewFormatAndBits(format, actualBitsPerValue)
}

// Reader A read-only random access array of positive integers.
// lucene.internal
type Reader interface {
	// Get the long at the given index. Behavior is undefined for out-of-range indices.
	Get(index int) uint64

	// GetBulk Bulk get: read at least one and at most len longs starting from index into
	// arr[off:off+len] and return the actual number of values that have been read.
	GetBulk(index int, arr []uint64) int

	// Size Returns: the number of values.
	Size() int
}

func GetBulk(reader Reader, index int, arr []uint64) int {
	gets := Min(reader.Size()-index, len(arr))

	for i, o, end := index, 0, index+gets; i < end; {
		arr[o] = reader.Get(i)
		i++
		o++
	}
	return gets
}

// ReaderIterator Run-once iterator interface, to decode previously saved PackedInts.
type ReaderIterator interface {
}

// NullReader A PackedInts.Reader which has all its values equal to 0 (bitsPerValue = 0).
type NullReader struct {
}

// Return the number of blocks required to store size values on blockSize.
func numBlocks(size, blockSize int) (int, error) {
	add := 1
	if size%blockSize == 0 {
		add = 0
	}
	numBlocks := size/blockSize + add
	if numBlocks*blockSize < size {
		return 0, errors.New("size is too large for this block size")
	}
	return numBlocks, nil
}

// PackedIntsCopy Copy src[srcPos:srcPos+len] into dest[destPos:destPos+len] using at most mem bytes.
func PackedIntsCopy(src Reader, srcPos int, dest Mutable, destPos, size, mem int) {
	capacity := mem >> 3
	if capacity == 0 {
		for i := 0; i < size; i++ {
			dest.Set(destPos, src.Get(srcPos))
			destPos++
			srcPos++
		}
	}

	if size > 0 {
		// use bulk operations
		buf := make([]uint64, Min(capacity, size))
		PackedIntsCopyBuff(src, srcPos, dest, destPos, size, buf)
	}
}

func PackedIntsCopyBuff(src Reader, srcPos int, dest Mutable, destPos, size int, buf []uint64) {
	remaining := 0
	for size > 0 {
		read := src.GetBulk(srcPos, buf[0:Min(size, len(buf)-remaining)])
		srcPos += read
		size -= read
		remaining += read

		written := dest.SetBulk(destPos, buf[0:remaining])
		destPos += written
		if written < remaining {
			copy(buf[0:remaining-written], buf[written:remaining])
		}
		remaining -= written
	}

	for remaining > 0 {
		written := dest.SetBulk(destPos, buf[0:remaining])
		destPos += written
		remaining -= written
		copy(buf[0:remaining], buf[written:written+remaining])
	}
}

// PackedIntsGetMutable Create a packed integer array with the given amount of values initialized to 0.
// the valueCount and the bitsPerValue cannot be changed after creation. All Mutables known by this
// factory are kept fully in RAM.
// Positive values of acceptableOverheadRatio will trade space for speed by selecting a faster but
// potentially less memory-efficient implementation. An acceptableOverheadRatio of COMPACT will make
// sure that the most memory-efficient implementation is selected whereas FASTEST will make sure that
// the fastest implementation is selected.
//
// Params: 	valueCount – the number of elements
//
//	bitsPerValue – the number of bits available for any given value
//	acceptableOverheadRatio – an acceptable overhead ratio per value
//
// Returns: a mutable packed integer array
// lucene.internal
func PackedIntsGetMutable(valueCount, bitsPerValue int, acceptableOverheadRatio float64) Mutable {
	formatAndBits := fastestFormatAndBits(valueCount, bitsPerValue, acceptableOverheadRatio)
	return getMutable(valueCount, bitsPerValue, formatAndBits.format)
}

func getMutableV1(valueCount, bitsPerValue int, acceptableOverheadRatio float64) Mutable {
	formatAndBits := fastestFormatAndBits(valueCount, bitsPerValue, acceptableOverheadRatio)
	return getMutable(valueCount, bitsPerValue, formatAndBits.format)
}

// Same as getMutable(int, int, float) with a pre-computed number of bits per value and format.
// lucene.internal
func getMutable(valueCount, bitsPerValue int, format Format) Mutable {
	switch format.(type) {
	case *formatPackedSingleBlock:
		return Packed64SingleBlockCreate(valueCount, bitsPerValue)
	case *formatPacked:
		switch bitsPerValue {
		case 8:
			return NewDirect8(valueCount)
		case 16:
			return NewDirect16(valueCount)
		case 32:
			return NewDirect32(valueCount)
		case 64:
			return NewDirect64(valueCount)
		case 24:
			if valueCount <= Packed8ThreeBlocksMaxSize {
				return NewPacked8ThreeBlocks(valueCount)
			}
			break
		case 48:
			if valueCount <= Packed16ThreeBlocksMaxSize {
				return NewPacked16ThreeBlocks(valueCount)
			}
			break
		}
		return NewPacked64(valueCount, bitsPerValue)
	default:
		//throw new AssertionError();
		return nil
	}
}

func getWriterNoHeader(out store.DataOutput, format Format, valueCount, bitsPerValue, mem int) Writer {
	return NewPackedWriter(format, out, valueCount, bitsPerValue, mem)
}

/**
  public static long maxValue(int bitsPerValue) {
    return bitsPerValue == 64 ? Long.MAX_VALUE : ~(~0L << bitsPerValue);
  }
*/

func PackedIntsMaxValue(bitsPerValue int) uint64 {
	if bitsPerValue == 64 {
		return math.MaxInt64
	}
	return ^(^uint64(0) << bitsPerValue)
}

// PackedIntsBitsRequired Returns how many bits are required to hold values up to and including maxValue
// NOTE: This method returns at least 1.
// Params: maxValue – the maximum value that should be representable.
// Returns: the amount of bits needed to represent values from 0 to maxValue.
// lucene.internal
func PackedIntsBitsRequired(maxValue uint64) int {
	return unsignedBitsRequired(maxValue)
}

func unsignedBitsRequired(v uint64) int {
	return Max(1, 64-bits.LeadingZeros64(v))
}

func checkBlockSize(blockSize, minBlockSize, maxBlockSize int) int {
	return bits.LeadingZeros32(uint32(blockSize))
}
