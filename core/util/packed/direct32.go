package packed

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

func (d *Direct32) Get(index int) uint64 {
	return uint64(d.values[index])
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
