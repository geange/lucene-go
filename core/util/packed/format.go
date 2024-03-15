package packed

import (
	"fmt"
	"math"
)

var (
	FormatPacked            = newFormatPacked(0)
	FormatPackedSingleBlock = newFormatPackedSingleBlock(1)
)

func GetFormatById(id int) (Format, error) {
	switch id {
	case 0:
		return FormatPacked, nil
	case 1:
		return FormatPackedSingleBlock, nil
	default:
		return nil, fmt.Errorf("unknown format id: %d", id)
	}
}

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

var _ Format = &intsFormat{}

type intsFormat struct {
	id int
}

func (f *intsFormat) GetId() int {
	return f.id
}

func (f *intsFormat) ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	if bitsPerValue >= 0 && bitsPerValue <= 64 {
		return bitsPerValue
	}

	return 7 * f.LongCount(packedIntsVersion, valueCount, bitsPerValue)
}

func (f *intsFormat) LongCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	if bitsPerValue >= 0 && bitsPerValue <= 64 {
		return bitsPerValue
	}

	byteCount := f.ByteCount(packedIntsVersion, valueCount, bitsPerValue)

	if byteCount%8 == 0 {
		return byteCount / 8
	}
	return byteCount/8 + 1
}

func (f *intsFormat) IsSupported(bitsPerValue int) bool {
	return bitsPerValue >= 1 && bitsPerValue <= 64
}

func (f *intsFormat) OverheadPerValue(bitsPerValue int) float64 {
	return 0
}

func (f *intsFormat) OverheadRatio(bitsPerValue int) float64 {
	return f.OverheadPerValue(bitsPerValue) / float64(bitsPerValue)
}

var _ Format = &formatPacked{}

type formatPacked struct {
	format *intsFormat
}

func newFormatPacked(id int) *formatPacked {
	return &formatPacked{format: &intsFormat{id: id}}
}

func (f *formatPacked) GetId() int {
	return f.format.GetId()
}

func (f *formatPacked) ByteCount(_packedIntsVersion, valueCount, bitsPerValue int) int {
	return int(math.Ceil(float64(valueCount*bitsPerValue) / 8))
}

func (f *formatPacked) LongCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	byteCount := f.ByteCount(packedIntsVersion, valueCount, bitsPerValue)

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
	format *intsFormat
}

func newFormatPackedSingleBlock(id int) *formatPackedSingleBlock {
	return &formatPackedSingleBlock{format: &intsFormat{id: id}}
}

func (f *formatPackedSingleBlock) GetId() int {
	return f.format.GetId()
}

func (f *formatPackedSingleBlock) ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	return 8 * f.LongCount(packedIntsVersion, valueCount, bitsPerValue)
}

func (f *formatPackedSingleBlock) LongCount(_packedIntsVersion, valueCount, bitsPerValue int) int {
	valuesPerBlock := 64 / bitsPerValue
	return int(math.Ceil(float64(valueCount) / float64(valuesPerBlock)))
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
