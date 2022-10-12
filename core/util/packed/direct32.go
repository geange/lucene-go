package packed

import (
	. "github.com/geange/lucene-go/math"
)

var _ Mutable = &Direct32{}

type Direct32 struct {
	*MutableImpl

	values []uint32
}

func NewDirect32(valueCount int) *Direct32 {
	direct := &Direct32{values: make([]uint32, valueCount)}
	direct.MutableImpl = newMutableImpl(direct, valueCount, 32)
	return direct
}

func (d *Direct32) Get(index int) uint64 {
	return uint64(d.values[index] & 0xFFFFFFFF)
}

func (d *Direct32) Set(index int, value uint64) {
	d.values[index] = uint32(value)
}

func (d *Direct32) Clear() {
	for i := range d.values {
		d.values[i] = 0
	}
}

func (d *Direct32) GetBulk(index int, arr []uint64) int {
	gets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = uint64(d.values[index+i] & 0xFFFFFFFF)
	}
	return gets
}

func (d *Direct32) SetBulk(index int, arr []uint64) int {
	sets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		d.values[index+i] = uint32(arr[i])
	}
	return sets
}

func (d *Direct32) Fill(fromIndex, toIndex int, value uint64) {
	for i := fromIndex; i < toIndex; i++ {
		d.values[i] = uint32(value)
	}
}
