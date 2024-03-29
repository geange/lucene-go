package packed

import (
	"context"

	"github.com/geange/lucene-go/core/store"
)

var _ Mutable = &GrowableWriter{}

type FuncGetMutable func(startBitsPerValue, valueCount int, acceptableOverheadRatio float64) Mutable

// GrowableWriter
// Implements Mutable, but grows the bit count of the underlying packed ints on-demand.
// Beware that this class will accept to set negative values but in order to do this,
// it will grow the number of bits per value to 64.
// @lucene.internal
type GrowableWriter struct {
	currentMask             uint64
	current                 Mutable
	acceptableOverheadRatio float64
	funcGetMutable          FuncGetMutable
}

// NewGrowableWriter
// startBitsPerValue: the initial number of bits per value, may grow depending on the data
// valueCount: the number of values
// acceptableOverheadRatio: an acceptable overhead ratio
func NewGrowableWriter(startBitsPerValue, valueCount int, acceptableOverheadRatio float64) *GrowableWriter {
	res := &GrowableWriter{
		acceptableOverheadRatio: acceptableOverheadRatio,
		funcGetMutable:          DefaultGetMutable,
	}
	res.current = res.funcGetMutable(valueCount, startBitsPerValue, acceptableOverheadRatio)
	return res
}

func getMask(bitsPerValue int) uint64 {
	if bitsPerValue == 64 {
		return ^uint64(0)
	}
	return MaxValue(bitsPerValue)
}

func (g *GrowableWriter) Get(index int) (uint64, error) {
	return g.current.Get(index)
}

func (g *GrowableWriter) GetTest(index int) uint64 {
	v, _ := g.Get(index)
	return v
}

func (g *GrowableWriter) GetBulk(index int, arr []uint64) int {
	return g.current.GetBulk(index, arr)
}

func (g *GrowableWriter) Size() int {
	return g.current.Size()
}

func (g *GrowableWriter) GetBitsPerValue() int {
	return g.current.GetBitsPerValue()
}

func (g *GrowableWriter) Set(index int, value uint64) {
	g.ensureCapacity(value)
	g.current.Set(index, value)
}

func (g *GrowableWriter) GetMutable() Mutable {
	return g.current
}

func (g *GrowableWriter) ensureCapacity(value uint64) {
	if value&g.currentMask == value {
		return
	}

	bitsRequired := unsignedBitsRequired(value)
	valueCount := g.Size()
	next := g.funcGetMutable(valueCount, bitsRequired, g.acceptableOverheadRatio)
	CopyValues(g.current, 0, next, 0, valueCount, DEFAULT_BUFFER_SIZE)
	g.current = next
	g.currentMask = getMask(g.current.GetBitsPerValue())
}

func (g *GrowableWriter) SetBulk(index int, values []uint64) int {
	maxCap := uint64(0)
	for i := 0; i < len(values); i++ {
		// bitwise or is nice because either all values are positive and the
		// or-ed result will require as many bits per value as the max of the
		// values, or one of them is negative and the result will be negative,
		// forcing GrowableWriter to use 64 bits per value
		maxCap |= values[i]
	}
	g.ensureCapacity(maxCap)
	return g.current.SetBulk(index, values)
}

func (g *GrowableWriter) Fill(fromIndex, toIndex int, value uint64) {
	g.ensureCapacity(value)
	g.current.Fill(fromIndex, toIndex, value)
}

func (g *GrowableWriter) Clear() {
	g.current.Clear()
}

func (g *GrowableWriter) Save(ctx context.Context, out store.DataOutput) error {
	return g.current.Save(ctx, out)
}

func (g *GrowableWriter) GetFormat() Format {
	return g.current.GetFormat()
}

func (g *GrowableWriter) Resize(newSize int) *GrowableWriter {
	next := NewGrowableWriter(g.GetBitsPerValue(), newSize, g.acceptableOverheadRatio)
	limit := min(g.Size(), newSize)
	CopyValues(g.current, 0, next, 0, limit, DEFAULT_BUFFER_SIZE)
	return next
}
