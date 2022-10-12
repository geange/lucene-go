package packed

import (
	. "github.com/geange/lucene-go/math"
)

var _ Mutable = &Direct16{}

type Direct16 struct {
	*MutableImpl

	values []uint16
}

func NewDirect16(valueCount int) *Direct16 {
	direct := &Direct16{values: make([]uint16, valueCount)}
	direct.MutableImpl = newMutableImpl(direct, valueCount, 16)
	return direct
}

func (d *Direct16) Get(index int) int64 {
	return int64(d.values[index] & 0xFFFF)
}

func (d *Direct16) Set(index int, value int64) {
	d.values[index] = uint16(value)
}

func (d *Direct16) Clear() {
	for i := range d.values {
		d.values[i] = 0
	}
}

func (d *Direct16) GetBulk(index int, arr []int64) int {
	gets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = int64(d.values[index+i] & 0xFFFF)
	}
	return gets
}

func (d *Direct16) SetBulk(index int, arr []int64) int {
	sets := Min(d.valueCount-index, len(arr))
	for i := range arr {
		d.values[index+i] = uint16(arr[i])
	}
	return sets
}

func (d *Direct16) Fill(fromIndex, toIndex int, value int64) {
	for i := fromIndex; i < toIndex; i++ {
		d.values[i] = uint16(value)
	}
}
