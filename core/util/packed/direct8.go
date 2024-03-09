package packed

import "github.com/geange/lucene-go/core/store"

var _ Mutable = &Direct8{}

// Direct8
// Direct wrapping of 8-bits values to a backing array.
// lucene.internal
type Direct8 struct {
	*BaseMutable

	values []byte
}

func NewDirect8(valueCount int) *Direct8 {
	direct := &Direct8{values: make([]byte, valueCount)}
	direct.BaseMutable = newBaseMutable(direct, valueCount, 8)
	return direct
}

func NewDirect8V1(packedIntsVersion int, in store.DataInput, valueCount int) (*Direct8, error) {
	direct := NewDirect8(valueCount)
	if _, err := in.Read(direct.values); err != nil {
		return nil, err
	}
	// because packed ints have not always been byte-aligned
	remaining := FormatPacked.ByteCount(packedIntsVersion, valueCount, 8) - 1*valueCount
	for i := 0; i < remaining; i++ {
		if _, err := in.ReadByte(); err != nil {
			return nil, err
		}
	}
	return direct, nil
}

func (d *Direct8) Get(index int) (uint64, error) {
	return uint64(d.values[index]), nil
}

func (d *Direct8) GetTest(index int) uint64 {
	v, _ := d.Get(index)
	return v
}

func (d *Direct8) Set(index int, value uint64) {
	d.values[index] = byte(value)
}

func (d *Direct8) Clear() {
	clear(d.values)
}

func (d *Direct8) GetBulk(index int, arr []uint64) int {
	gets := min(d.valueCount-index, len(arr))
	for i := range arr {
		arr[i] = uint64(d.values[index+i])
	}
	return gets
}

func (d *Direct8) SetBulk(index int, arr []uint64) int {
	sets := min(d.valueCount-index, len(arr))
	for i := range arr {
		d.values[index+i] = byte(arr[i])
	}
	return sets
}

func (d *Direct8) Fill(fromIndex, toIndex int, value uint64) {
	for i := fromIndex; i < toIndex; i++ {
		d.values[i] = byte(value)
	}
}
