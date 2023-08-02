package packed

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/bits"

	"github.com/geange/lucene-go/core/store"
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

	acceptableOverheadRatio = max(COMPACT, acceptableOverheadRatio)
	acceptableOverheadRatio = min(FASTEST, acceptableOverheadRatio)
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
	} else if valueCount <= Packed8ThreeBlocks_MAX_SIZE && bitsPerValue <= 24 && maxBitsPerValue >= 24 {
		actualBitsPerValue = 24
	} else if valueCount <= Packed8ThreeBlocks_MAX_SIZE && bitsPerValue <= 48 && maxBitsPerValue >= 48 {
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

func GetBulk(reader Reader, index int, arr []uint64) int {
	gets := min(reader.Size()-index, len(arr))

	for i, o, end := index, 0, index+gets; i < end; {
		arr[o], _ = reader.Get(i)
		i++
		o++
	}
	return gets
}

// ReaderIterator Run-once iterator interface, to decode previously saved PackedInts.
type ReaderIterator interface {
}

var _ Reader = &NullReader{}

// NullReader
// A PackedInts.PackedIntsReader which has all its values equal to 0 (bitsPerValue = 0).
type NullReader struct {
	valueCount int
}

func NewNullReader(valueCount int) *NullReader {
	return &NullReader{valueCount: valueCount}
}

func (n *NullReader) Get(index int) (uint64, error) {
	return 0, nil
}

func (n *NullReader) GetBulk(index int, arr []uint64) int {
	size := min(len(arr), n.valueCount-index)
	clear(arr[:size])
	return size
}

func (n *NullReader) Size() int {
	return n.valueCount
}

// Return the number of blocks required to store size values on blockSize.
func getNumBlocks(size, blockSize int) (int, error) {
	add := 1
	if size%blockSize == 0 {
		add = 0
	}
	num := size/blockSize + add
	if num*blockSize < size {
		return 0, errors.New("size is too large for this block size")
	}
	return num, nil
}

// CopyValues
// Copy src[srcPos:srcPos+len] into dest[destPos:destPos+len] using at most mem bytes.
func CopyValues(src Reader, srcPos int, dest Mutable, destPos, size, mem int) {
	capacity := mem >> 3
	if capacity == 0 {
		for i := 0; i < size; i++ {
			v, err := src.Get(srcPos)
			if err != nil {
				return
			}
			dest.Set(destPos, v)
			destPos++
			srcPos++
		}
	}

	if size > 0 {
		// use bulk operations
		buf := make([]uint64, min(capacity, size))
		CopyValuesWithBuffer(src, srcPos, dest, destPos, size, buf)
	}
}

// CopyValuesWithBuffer
// Same as copy(PackedInts.Reader, int, PackedInts.Mutable, int, int, int) but using a pre-allocated buffer.
func CopyValuesWithBuffer(src Reader, srcPos int, dest Mutable, destPos, size int, buf []uint64) {
	remaining := 0
	for size > 0 {
		copySize := min(size, len(buf)-remaining)
		read := src.GetBulk(srcPos, buf[0:copySize])
		srcPos += read
		size -= read
		remaining += read

		written := dest.SetBulk(destPos, buf[0:remaining])
		destPos += written
		if written < remaining {
			length := remaining - written
			copy(buf[:length], buf[written:])
		}
		remaining -= written
	}

	for remaining > 0 {
		written := dest.SetBulk(destPos, buf[0:remaining])
		destPos += written
		remaining -= written
		copy(buf[0:remaining], buf[written:])
	}
}

// DefaultGetMutable
// Create a packed integer array with the given amount of values initialized to 0.
// the valueCount and the bitsPerValue cannot be changed after creation. All Mutables known by this
// factory are kept fully in RAM.
// Positive values of acceptableOverheadRatio will trade space for speed by selecting a faster but
// potentially less memory-efficient implementation. An acceptableOverheadRatio of COMPACT will make
// sure that the most memory-efficient implementation is selected whereas FASTEST will make sure that
// the fastest implementation is selected.
//
// valueCount: the number of elements
// bitsPerValue: the number of bits available for any given value
// acceptableOverheadRatio: an acceptable overhead ratio per value
//
// Returns a mutable packed integer array
// lucene.internal
func DefaultGetMutable(valueCount, bitsPerValue int, acceptableOverheadRatio float64) Mutable {
	return getMutableV1(valueCount, bitsPerValue, acceptableOverheadRatio)
}

func getMutableV1(valueCount, bitsPerValue int, acceptableOverheadRatio float64) Mutable {
	formatAndBits := fastestFormatAndBits(valueCount, bitsPerValue, acceptableOverheadRatio)
	return getMutable(valueCount, bitsPerValue, formatAndBits.format)
}

// Same as getMutable(int, int, float) with a pre-computed number of bits per value and intsFormat.
// lucene.internal
func getMutable(valueCount, bitsPerValue int, format Format) Mutable {
	switch format.(type) {
	case *formatPackedSingleBlock:
		m, _ := NewPacked64SingleBlock(valueCount, bitsPerValue)
		return m
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
			if valueCount <= Packed8ThreeBlocks_MAX_SIZE {
				return NewPacked8ThreeBlocks(valueCount)
			}
			break
		case 48:
			if valueCount <= Packed16ThreeBlocks_MAX_SIZE {
				return NewPacked16ThreeBlocks(valueCount)
			}
			break
		}
		return NewPacked64(valueCount, bitsPerValue)
	default:
		return nil
	}
}

func getWriterNoHeader(out store.DataOutput, format Format, valueCount, bitsPerValue, mem int) Writer {
	return NewPackedWriter(format, out, valueCount, bitsPerValue, mem)
}

func MaxValue(bitsPerValue int) uint64 {
	if bitsPerValue == 64 {
		return math.MaxInt64
	}
	return ^(^uint64(0) << bitsPerValue)
}

// BitsRequired
// Returns how many bits are required to hold values up to and including maxValue
// NOTE: This method returns at least 1.
// Params: maxValue – the maximum value that should be representable.
// Returns: the amount of bits needed to represent values from 0 to maxValue.
// lucene.internal
func BitsRequired(maxValue int64) (int, error) {
	if maxValue < 0 {
		return 0, fmt.Errorf("maxValue must be non-negative (got: %d)", maxValue)
	}
	return unsignedBitsRequired(uint64(maxValue)), nil
}

func UnsignedBitsRequired(v uint64) int {
	return unsignedBitsRequired(v)
}

func unsignedBitsRequired(v uint64) int {
	return max(1, 64-bits.LeadingZeros64(v))
}

func checkBlockSize(blockSize, minBlockSize, maxBlockSize int) int {
	return bits.TrailingZeros(uint(blockSize))
}

// Decoder A decoder for packed integers.
type Decoder interface {

	// LongBlockCount
	// The minimum number of long blocks to encode in a single iteration, when using long encoding.
	LongBlockCount() int

	// LongValueCount
	// The number of values that can be stored in longBlockCount() long blocks.
	LongValueCount() int

	// ByteBlockCount
	// The minimum number of byte blocks to encode in a single iteration, when using byte encoding.
	ByteBlockCount() int

	// ByteValueCount
	// The number of values that can be stored in byteBlockCount() byte blocks.
	ByteValueCount() int

	// DecodeUint64
	// Read iterations * blockCount() blocks from blocks,
	// decode them and write iterations * valueCount() values into values.
	// blocks: the long blocks that hold packed integer values
	// values: the values buffer
	// iterations: controls how much data to decode
	DecodeUint64(blocks []uint64, values []uint64, iterations int)

	// DecodeBytes
	// Read 8 * iterations * blockCount() blocks from blocks,
	// decode them and write iterations * valueCount() values into values.
	// blocks: the long blocks that hold packed integer values
	// values: the values buffer
	// iterations: controls how much data to decode
	DecodeBytes(blocks []byte, values []uint64, iterations int)
}

// Encoder An encoder for packed integers.
type Encoder interface {
	// LongBlockCount
	// The minimum number of long blocks to encode in a single iteration, when using long encoding.
	LongBlockCount() int

	// LongValueCount
	// The number of values that can be stored in longBlockCount() long blocks.
	LongValueCount() int

	// ByteBlockCount
	// The minimum number of byte blocks to encode in a single iteration, when using byte encoding.
	ByteBlockCount() int

	// ByteValueCount
	// The number of values that can be stored in byteBlockCount() byte blocks.
	ByteValueCount() int

	// EncodeUint64
	// Read iterations * valueCount() values from values, encode them and write
	// iterations * blockCount() blocks into blocks.
	// values: the values buffer
	// blocks: the long blocks that hold packed integer values
	// iterations: controls how much data to encode
	EncodeUint64(values []uint64, blocks []uint64, iterations int)

	// EncodeBytes
	// Read iterations * valueCount() values from values,
	// encode them and write 8 * iterations * blockCount() blocks into blocks.
	// values: the values buffer
	// blocks: the long blocks that hold packed integer values
	// iterations: controls how much data to encode
	EncodeBytes(values []uint64, blocks []byte, iterations int)
}

// GetEncoder
// Get an PackedInts.Encoder.
// format: the format used to store packed ints version – the compatibility version bitsPerValue – the number of bits per value
func GetEncoder(format Format, version, bitsPerValue int) (Encoder, error) {
	if err := checkVersion(version); err != nil {
		return nil, err
	}

	return Of(format, bitsPerValue)
}

func GetDecoder(format Format, version, bitsPerValue int) (Decoder, error) {
	if err := checkVersion(version); err != nil {
		return nil, err
	}
	return Of(format, bitsPerValue)
}

func checkVersion(version int) error {
	if version < VERSION_START {
		return errors.New("version is too old")
	}
	if version > VERSION_CURRENT {
		return errors.New("version is too new")
	}
	return nil
}

// Expert: Restore a PackedInts.Reader from a stream without reading metadata at the beginning of the stream.
// This method is useful to restore data from streams which have been created using
// getWriterNoHeader(store.DataOutput, Format, int, int, int).
//
// in: the stream to read data from, positioned at the beginning of the packed values
// format: the format used to serialize
// version: the version used to serialize the data
// valueCount: how many values the stream holds
// bitsPerValue: the number of bits per value
//
// lucene.internal
func getReaderNoHeader(ctx context.Context, in store.IndexInput, format Format, version, valueCount, bitsPerValue int) (Reader, error) {
	if err := checkVersion(version); err != nil {
		return nil, err
	}

	switch format.(type) {
	case *formatPackedSingleBlock:
		return NewPacked64SingleBlockV1(nil, in, valueCount, bitsPerValue)
	case *formatPacked:
		switch bitsPerValue {
		case 8:
			return NewDirect8V1(ctx, version, in, valueCount)
		case 16:
			return NewDirect16V1(ctx, version, in, valueCount)
		case 32:
			return NewDirect32V1(ctx, version, in, valueCount)
		case 64:
			return NewDirect64V1(ctx, version, in, valueCount)
		case 24:
			if valueCount <= Packed8ThreeBlocks_MAX_SIZE {
				return NewNewPacked8ThreeBlocksV1(ctx, version, in, valueCount)
			}
		case 48:
			if valueCount <= Packed16ThreeBlocks_MAX_SIZE {
				return NewPacked16ThreeBlocksV1(ctx, version, in, valueCount)
			}
		}
		return NewPacked64V1(ctx, version, in, valueCount, bitsPerValue)
	default:
		return nil, errors.New("unknown Writer format")
	}
}

// Expert: Construct a direct Reader from a stream without reading
// metadata at the beginning of the stream. This method is useful to restore
// data from streams which have been created using getWriterNoHeader(store.DataOutput, Format, int, int, int)
//
// The returned reader will have very little memory overhead, but every call
// to {@link Reader#get(int)} is likely to perform a disk seek.
//
// in: the stream to read data from
// format: the format used to serialize
// version: the version used to serialize the data
// valueCount: how many values the stream holds
// bitsPerValue: the number of bits per value
// @lucene.internal
func getDirectReaderNoHeader(ctx context.Context, in store.IndexInput, format Format, version, valueCount, bitsPerValue int) (Reader, error) {
	err := checkVersion(version)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatPacked:
		return NewDirectPackedReader(bitsPerValue, valueCount, in), nil
	case FormatPackedSingleBlock:
		return NewDirectPacked64SingleBlockReader(bitsPerValue, valueCount, in), nil
	default:
		return nil, errors.New("unknown format")
	}
}
