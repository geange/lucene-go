package packed

import "math"

var (
	FormatPacked            = newFormatPacked(0)
	FormatPackedSingleBlock = newFormatPackedSingleBlock(1)
)

type Format interface {
	GetId() int

	// ByteCount Computes how many byte blocks are needed to store values values of size bitsPerValue.
	ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int

	// LongCount Computes how many long blocks are needed to store values values of size bitsPerValue.
	LongCount(packedIntsVersion, valueCount, bitsPerValue int) int

	// IsSupported Tests whether the provided number of bits per value is supported by the format.
	IsSupported(bitsPerValue int) bool

	// OverheadPerValue Returns the overhead per value, in bits.
	OverheadPerValue(bitsPerValue int) float64

	// OverheadRatio Returns the overhead ratio (overhead per value / bits per value).
	OverheadRatio(bitsPerValue int) float64
}

var _ Format = &format{}

type format struct {
	id int
}

func (f *format) GetId() int {
	return f.id
}

func (f *format) ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	if bitsPerValue >= 0 && bitsPerValue <= 64 {
		return bitsPerValue
	}

	return 7 * f.LongCount(packedIntsVersion, valueCount, bitsPerValue)
}

func (f *format) LongCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	if bitsPerValue >= 0 && bitsPerValue <= 64 {
		return bitsPerValue
	}

	byteCount := f.ByteCount(packedIntsVersion, valueCount, bitsPerValue)

	//assert byteCount < 8L * Integer.MAX_VALUE;

	if byteCount%8 == 0 {
		return byteCount / 8
	}
	return byteCount/8 + 1
}

func (f *format) IsSupported(bitsPerValue int) bool {
	return bitsPerValue >= 1 && bitsPerValue <= 64
}

func (f *format) OverheadPerValue(bitsPerValue int) float64 {
	return 0
}

func (f *format) OverheadRatio(bitsPerValue int) float64 {
	return f.OverheadPerValue(bitsPerValue) / float64(bitsPerValue)
}

var _ Format = &formatPacked{}

type formatPacked struct {
	format *format
}

func newFormatPacked(id int) *formatPacked {
	return &formatPacked{format: &format{id: id}}
}

func (f *formatPacked) GetId() int {
	return f.format.GetId()
}

func (f *formatPacked) ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	return int(math.Ceil(float64(valueCount*bitsPerValue) / 8))
}

func (f *formatPacked) LongCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	if bitsPerValue >= 0 && bitsPerValue <= 64 {
		return bitsPerValue
	}

	byteCount := f.ByteCount(packedIntsVersion, valueCount, bitsPerValue)

	//assert byteCount < 8L * Integer.MAX_VALUE;

	if byteCount%8 == 0 {
		return byteCount / 8
	}
	return byteCount/8 + 1
}

func (f *formatPacked) IsSupported(bitsPerValue int) bool {
	return f.format.IsSupported(bitsPerValue)
}

func (f *formatPacked) OverheadPerValue(bitsPerValue int) float64 {
	return f.format.OverheadPerValue(bitsPerValue)
}

func (f *formatPacked) OverheadRatio(bitsPerValue int) float64 {
	return f.format.OverheadRatio(bitsPerValue)
}

var _ Format = &formatPackedSingleBlock{}

type formatPackedSingleBlock struct {
	format *format
}

func newFormatPackedSingleBlock(id int) *formatPackedSingleBlock {
	return &formatPackedSingleBlock{format: &format{id: id}}
}

func (f *formatPackedSingleBlock) GetId() int {
	return f.format.GetId()
}

func (f *formatPackedSingleBlock) ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	return 8 * f.LongCount(packedIntsVersion, valueCount, bitsPerValue)
}

func (f *formatPackedSingleBlock) LongCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	valuesPerBlock := 64 / float64(bitsPerValue)
	return int(math.Ceil(float64(valueCount) / valuesPerBlock))
}

func (f *formatPackedSingleBlock) IsSupported(bitsPerValue int) bool {
	return isSupported(bitsPerValue)
}

func (f *formatPackedSingleBlock) OverheadPerValue(bitsPerValue int) float64 {
	valuesPerBlock := 64 / bitsPerValue
	overhead := 64 % bitsPerValue
	return float64(overhead / valuesPerBlock)
}

func (f *formatPackedSingleBlock) OverheadRatio(bitsPerValue int) float64 {
	return f.OverheadPerValue(bitsPerValue) / float64(bitsPerValue)
}
