package packed

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

func (d *Direct16) Get(index int) uint64 {
	return uint64(d.values[index])
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
