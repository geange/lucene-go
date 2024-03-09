package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
)

var _ Mutable = &Direct64{}

type Direct64 struct {
	*BaseMutable
	values []uint64
}

func NewDirect64(valueCount int) *Direct64 {
	direct := &Direct64{values: make([]uint64, valueCount)}
	direct.BaseMutable = newBaseMutable(direct, valueCount, 64)
	return direct
}

func NewDirect64V1(packedIntsVersion int, in store.DataInput, valueCount int) (*Direct64, error) {
	direct := NewDirect64(valueCount)

	for i := 0; i < valueCount; i++ {
		num, err := in.ReadUint64(context.TODO())
		if err != nil {
			return nil, err
		}
		direct.values[i] = num
	}

	// because packed ints have not always been byte-aligned
	remaining := FormatPacked.ByteCount(packedIntsVersion, valueCount, 64) - 8*valueCount
	for i := 0; i < remaining; i++ {
		if _, err := in.ReadByte(); err != nil {
			return nil, err
		}
	}
	return direct, nil
}

func (d *Direct64) Get(index int) (uint64, error) {
	return d.values[index], nil
}

func (d *Direct64) GetTest(index int) uint64 {
	v, _ := d.Get(index)
	return v
}

func (d *Direct64) Set(index int, value uint64) {
	d.values[index] = value
}

func (d *Direct64) Clear() {
	clear(d.values)
}

func (d *Direct64) GetBulk(index int, arr []uint64) int {
	gets := min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = d.values[index+i]
	}
	return gets
}

func (d *Direct64) SetBulk(index int, arr []uint64) int {
	sets := min(d.valueCount-index, len(arr))
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
