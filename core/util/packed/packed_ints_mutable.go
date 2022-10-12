package packed

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
)

// Mutable A packed integer array that can be modified.
// lucene.internal
type Mutable interface {
	Reader

	// GetBitsPerValue Returns:
	//the number of bits used to store any given value. Note: This does not imply that memory usage is bitsPerValue * #values as implementations are free to use non-space-optimal packing of bits.
	GetBitsPerValue() int

	// Set the value at the given index in the array.
	// Params:  index – where the value should be positioned.
	//			value – a value conforming to the constraints set by the array.
	Set(index int, value int64)

	// SetBulk Bulk set: set at least one and at most len longs starting at off in arr into this mutable, starting at index. Returns the actual number of values that have been set.
	SetBulk(index int, arr []int64) int

	// Fill the mutable from fromIndex (inclusive) to toIndex (exclusive) with val.
	Fill(fromIndex, toIndex int, value int64)

	// Clear Sets all values to 0.
	Clear()

	// Save this mutable into out. Instantiating a reader from the generated data will return a
	// reader with the same number of bits per value.
	Save(out store.DataOutput) error

	GetFormat() Format
}

type mutableSpi interface {
	GetBitsPerValue() int
	Get(index int) int64
	Size() int
	Set(index int, value int64)
}

type mutable struct {
	spi mutableSpi
}

func (m *mutable) GetBulk(index int, arr []int64) int {
	length := len(arr)
	gets := Min(m.spi.Size()-index, length)

	for i, o, end := index, 0, index+gets; i < end; i++ {
		arr[o] = m.spi.Get(i)
		i++
		o++
	}

	return gets
}

func (m *mutable) SetBulk(index int, arr []int64) int {
	size := Min(len(arr), m.spi.Size()-index)

	for i, o, end := index, 0, index+len(arr); i < end; {
		m.spi.Set(i, arr[o])
		i++
		o++
	}
	return size
}

func (m *mutable) Fill(fromIndex, toIndex int, value int64) {
	for i := fromIndex; i < toIndex; i++ {
		m.spi.Set(i, value)
	}
}

func (m *mutable) Clear() {
	m.Fill(0, m.spi.Size(), 0)
}

func (m *mutable) Save(out store.DataOutput) error {
	writer := getWriterNoHeader(out, m.GetFormat(), m.spi.Size(), m.spi.GetBitsPerValue(), DEFAULT_BUFFER_SIZE)
	err := writer.WriteHeader()
	if err != nil {
		return err
	}
	for i := 0; i < m.spi.Size(); i++ {
		err := writer.Add(m.spi.Get(i))
		if err != nil {
			return err
		}
	}
	return writer.Finish()
}

func (m *mutable) GetFormat() Format {
	return FormatPacked
}

type MutableImpl struct {
	*mutable

	valueCount   int
	bitsPerValue int
}

func newMutableImpl(spi mutableSpi, valueCount int, bitsPerValue int) *MutableImpl {
	return &MutableImpl{
		mutable:      &mutable{spi: spi},
		valueCount:   valueCount,
		bitsPerValue: bitsPerValue,
	}
}

func (m *MutableImpl) GetBitsPerValue() int {
	return m.bitsPerValue
}

func (m *MutableImpl) Size() int {
	return m.valueCount
}

func Fill(spi mutableSpi, fromIndex, toIndex int, value int64) {
	for i := fromIndex; i < toIndex; i++ {
		spi.Set(i, value)
	}
}

func SetBulk(spi mutableSpi, index int, arr []int64) int {
	size := Min(len(arr), spi.Size()-index)

	for i, o, end := index, 0, len(arr); i < end; {
		spi.Set(i, arr[o])
		i++
		o++
	}
	return size
}
