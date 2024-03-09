package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
)

var _ Mutable = &Direct16{}

type Direct16 struct {
	*BaseMutable

	values []uint16
}

func NewDirect16(valueCount int) *Direct16 {
	direct := &Direct16{values: make([]uint16, valueCount)}
	direct.BaseMutable = newBaseMutable(direct, valueCount, 16)
	return direct
}

func NewDirect16V1(packedIntsVersion int, in store.DataInput, valueCount int) (*Direct16, error) {
	direct := NewDirect16(valueCount)

	for i := 0; i < valueCount; i++ {
		num, err := in.ReadUint16(context.TODO())
		if err != nil {
			return nil, err
		}
		direct.values[i] = num
	}

	// because packed ints have not always been byte-aligned
	remaining := FormatPacked.ByteCount(packedIntsVersion, valueCount, 16) - 2*valueCount
	for i := 0; i < remaining; i++ {
		if _, err := in.ReadByte(); err != nil {
			return nil, err
		}
	}
	return direct, nil
}

func (d *Direct16) Get(index int) (uint64, error) {
	return uint64(d.values[index]), nil
}

func (d *Direct16) GetTest(index int) uint64 {
	v, _ := d.Get(index)
	return v
}

func (d *Direct16) Set(index int, value uint64) {
	d.values[index] = uint16(value)
}

func (d *Direct16) Clear() {
	clear(d.values)
}

func (d *Direct16) GetBulk(index int, arr []uint64) int {
	gets := min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = uint64(d.values[index+i] & 0xFFFF)
	}
	return gets
}

func (d *Direct16) SetBulk(index int, arr []uint64) int {
	sets := min(d.valueCount-index, len(arr))
	for i := range arr {
		d.values[index+i] = uint16(arr[i])
	}
	return sets
}

func (d *Direct16) Fill(fromIndex, toIndex int, value uint64) {
	for i := fromIndex; i < toIndex; i++ {
		d.values[i] = uint16(value)
	}
}
