package packed

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"io"
)

// Mutable
// A packed integer array that can be modified.
// lucene.internal
type Mutable interface {
	Reader

	// GetBitsPerValue returns the number of bits used to store any given value.
	// Note: This does not imply that memory usage is bitsPerValue * #values as implementations
	// are free to use non-space-optimal packing of bits.
	GetBitsPerValue() int

	// Set the value at the given index in the array.
	// index: where the value should be positioned.
	// value: a value conforming to the constraints set by the array.
	Set(index int, value uint64)

	// SetBulk set at least one and at most len longs starting at off in arr into this mutable,
	// starting at index. Returns the actual number of values that have been set.
	SetBulk(index int, arr []uint64) int

	// Fill the mutable from fromIndex (inclusive) to toIndex (exclusive) with val.
	Fill(fromIndex, toIndex int, value uint64)

	// Clear Sets all values to 0.
	Clear()

	// Save this mutable into out. Instantiating a reader from the generated data will return a
	// reader with the same number of bits per value.
	Save(ctx context.Context, out store.DataOutput) error

	GetFormat() Format
}

type mutableSPI interface {
	GetBitsPerValue() int
	Get(index int) (uint64, error)
	Size() int
	Set(index int, value uint64)
}

type mutable struct {
	spi mutableSPI
}

func (m *mutable) GetBulk(index int, arr []uint64) int {
	length := len(arr)
	gets := min(m.spi.Size()-index, length)

	for i, o, end := index, 0, index+gets; i < end; {
		n, err := m.spi.Get(i)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return i
			}
			return 0
		}

		arr[o] = n
		i++
		o++
	}

	return gets
}

func (m *mutable) SetBulk(index int, arr []uint64) int {
	size := min(len(arr), m.spi.Size()-index)

	i := index
	off := 0
	end := index + len(arr)
	for i < end {
		m.spi.Set(i, arr[off])
		i++
		off++
	}
	return size
}

func (m *mutable) Fill(fromIndex, toIndex int, value uint64) {
	for i := fromIndex; i < toIndex; i++ {
		m.spi.Set(i, value)
	}
}

func (m *mutable) Clear() {
	m.Fill(0, m.spi.Size(), 0)
}

func (m *mutable) Save(ctx context.Context, out store.DataOutput) error {
	writer := getWriterNoHeader(out, m.GetFormat(), m.spi.Size(), m.spi.GetBitsPerValue(), DEFAULT_BUFFER_SIZE)
	err := writer.WriteHeader(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < m.spi.Size(); i++ {
		n, err := m.spi.Get(i)
		if err != nil {
			return err
		}
		if err := writer.Add(n); err != nil {
			return err
		}
	}
	return writer.Finish()
}

func (m *mutable) GetFormat() Format {
	return FormatPacked
}

type baseMutable struct {
	*mutable

	valueCount   int
	bitsPerValue int
}

func newBaseMutable(spi mutableSPI, valueCount int, bitsPerValue int) *baseMutable {
	return &baseMutable{
		mutable:      &mutable{spi: spi},
		valueCount:   valueCount,
		bitsPerValue: bitsPerValue,
	}
}

func (m *baseMutable) GetBitsPerValue() int {
	return m.bitsPerValue
}

func (m *baseMutable) Size() int {
	return m.valueCount
}

func Fill(spi mutableSPI, fromIndex, toIndex int, value uint64) {
	for i := fromIndex; i < toIndex; i++ {
		spi.Set(i, value)
	}
}

func SetBulk(spi mutableSPI, index int, arr []uint64) int {
	size := min(len(arr), spi.Size()-index)

	i := index
	off := 0
	end := index + len(arr)

	for i < end {
		spi.Set(i, arr[off])
		i++
		off++
	}
	return size
}
