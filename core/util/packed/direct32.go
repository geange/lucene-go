package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
)

var _ Mutable = &Direct32{}

type Direct32 struct {
	*BaseMutable

	values []uint32
}

func NewDirect32(valueCount int) *Direct32 {
	direct := &Direct32{values: make([]uint32, valueCount)}
	direct.BaseMutable = newBaseMutable(direct, valueCount, 32)
	return direct
}

func NewDirect32V1(packedIntsVersion int, in store.DataInput, valueCount int) (*Direct32, error) {
	direct := NewDirect32(valueCount)

	for i := 0; i < valueCount; i++ {
		num, err := in.ReadUint32(context.TODO())
		if err != nil {
			return nil, err
		}
		direct.values[i] = num
	}

	// because packed ints have not always been byte-aligned
	remaining := FormatPacked.ByteCount(packedIntsVersion, valueCount, 32) - 4*valueCount
	for i := 0; i < remaining; i++ {
		if _, err := in.ReadByte(); err != nil {
			return nil, err
		}
	}
	return direct, nil
}

func (d *Direct32) Get(index int) (uint64, error) {
	return uint64(d.values[index]), nil
}

func (d *Direct32) GetTest(index int) uint64 {
	v, _ := d.Get(index)
	return v
}

func (d *Direct32) Set(index int, value uint64) {
	d.values[index] = uint32(value)
}

func (d *Direct32) Clear() {
	clear(d.values)
}

func (d *Direct32) GetBulk(index int, arr []uint64) int {
	gets := min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = uint64(d.values[index+i] & 0xFFFFFFFF)
	}
	return gets
}

func (d *Direct32) SetBulk(index int, arr []uint64) int {
	sets := min(d.valueCount-index, len(arr))
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
