package packed

import (
	. "github.com/geange/lucene-go/math"
)

var _ Mutable = &Direct8{}

// Direct8 Direct wrapping of 8-bits values to a backing array.
// lucene.internal
type Direct8 struct {
	*MutableImpl

	values []byte
}

func NewDirect8(valueCount int) *Direct8 {
	direct := &Direct8{values: make([]byte, valueCount)}
	direct.MutableImpl = newMutableImpl(direct, valueCount, 8)
	return direct
}

func (d *Direct8) Get(index int) int64 {
	return int64(d.values[index] & 0xFF)
}

func (d *Direct8) Set(index int, value int64) {
	d.values[index] = byte(value)
}

func (d *Direct8) Clear() {
	for i := range d.values {
		d.values[i] = 0
	}
}

func (d *Direct8) GetBulk(index int, arr []int64) int {
	gets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = int64(d.values[index+i] & 0xFF)
	}
	return gets
}

func (d *Direct8) SetBulk(index int, arr []int64) int {
	sets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		d.values[index+i] = byte(arr[i])
	}
	return sets
}

func (d *Direct8) Fill(fromIndex, toIndex int, value int64) {
	for i := fromIndex; i < toIndex; i++ {
		d.values[i] = byte(value)
	}
}
