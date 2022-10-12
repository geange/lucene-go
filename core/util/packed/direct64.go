package packed

import (
	. "github.com/geange/lucene-go/math"
)

var _ Mutable = &Direct64{}

type Direct64 struct {
	*MutableImpl
	values []uint64
}

func NewDirect64(valueCount int) *Direct64 {
	direct := &Direct64{values: make([]uint64, valueCount)}
	direct.MutableImpl = newMutableImpl(direct, valueCount, 64)
	return direct
}

func (d *Direct64) Get(index int) uint64 {
	return d.values[index]
}

func (d *Direct64) Set(index int, value uint64) {
	d.values[index] = value
}

func (d *Direct64) Clear() {
	for i := range d.values {
		d.values[i] = 0
	}
}

func (d *Direct64) GetBulk(index int, arr []uint64) int {
	gets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = d.values[index+i]
	}
	return gets
}

func (d *Direct64) SetBulk(index int, arr []uint64) int {
	sets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		d.values[index+i] = arr[i]
	}
	return sets
}

func (d *Direct64) Fill(fromIndex, toIndex int, value uint64) {
	for i := fromIndex; i < toIndex; i++ {
		d.values[i] = value
	}
}
