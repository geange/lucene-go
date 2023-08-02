package packed

import (
	"context"

	"github.com/geange/lucene-go/core/store"
)

var _ Mutable = &Direct8{}

// Direct8
// Direct wrapping of 8-bits values to a backing array.
// lucene.internal
type Direct8 struct {
	*baseMutable

	values []byte
}

func NewDirect8(valueCount int) *Direct8 {
	direct := &Direct8{values: make([]byte, valueCount)}
	direct.baseMutable = newBaseMutable(direct, valueCount, 8)
	return direct
}

func NewDirect8V1(_ context.Context, packedIntsVersion int, in store.DataInput, valueCount int) (*Direct8, error) {
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

var _ Mutable = &Direct16{}

type Direct16 struct {
	*baseMutable

	values []uint16
}

func NewDirect16(valueCount int) *Direct16 {
	direct := &Direct16{values: make([]uint16, valueCount)}
	direct.baseMutable = newBaseMutable(direct, valueCount, 16)
	return direct
}

func NewDirect16V1(ctx context.Context, packedIntsVersion int, in store.DataInput, valueCount int) (*Direct16, error) {
	direct := NewDirect16(valueCount)

	for i := 0; i < valueCount; i++ {
		num, err := in.ReadUint16(ctx)
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

var _ Mutable = &Direct32{}

type Direct32 struct {
	*baseMutable

	values []uint32
}

func NewDirect32(valueCount int) *Direct32 {
	direct := &Direct32{values: make([]uint32, valueCount)}
	direct.baseMutable = newBaseMutable(direct, valueCount, 32)
	return direct
}

func NewDirect32V1(ctx context.Context, packedIntsVersion int, in store.DataInput, valueCount int) (*Direct32, error) {
	direct := NewDirect32(valueCount)

	for i := 0; i < valueCount; i++ {
		num, err := in.ReadUint32(ctx)
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

var _ Mutable = &Direct64{}

type Direct64 struct {
	*baseMutable
	values []uint64
}

func NewDirect64(valueCount int) *Direct64 {
	direct := &Direct64{values: make([]uint64, valueCount)}
	direct.baseMutable = newBaseMutable(direct, valueCount, 64)
	return direct
}

func NewDirect64V1(ctx context.Context, packedIntsVersion int, in store.DataInput, valueCount int) (*Direct64, error) {
	direct := NewDirect64(valueCount)

	for i := 0; i < valueCount; i++ {
		num, err := in.ReadUint64(ctx)
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
