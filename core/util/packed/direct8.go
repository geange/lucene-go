package packed

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

func (d *Direct8) Get(index int) uint64 {
	return uint64(d.values[index])
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
