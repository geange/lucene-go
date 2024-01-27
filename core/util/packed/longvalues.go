package packed

import "github.com/geange/lucene-go/core/types"

var _ types.LongValues = &LongValues{}

// LongValues
// Utility class to compress integers into a LongValues instance.
// TODO: need to reduce memory
type LongValues struct {
	values    []Reader
	pageShift int
	pageMask  int64
	size      int
}

func NewLongValues(values []Reader, pageShift int, pageMask int, size int) *LongValues {
	return &LongValues{values: values, pageShift: pageShift, pageMask: int64(pageMask), size: size}
}

func (p *LongValues) Size() int64 {
	return int64(p.size)
}

func (p *LongValues) Get(index int64) int64 {
	block := int(index >> p.pageShift)
	element := int(index & p.pageMask)
	return p.Load(block, element)
}

func (p *LongValues) Load(block int, element int) int64 {
	return int64(p.values[block].Get(element))
}

type LongValuesIterator struct {
	p   *LongValues
	pos int
}

// HasNext Whether or not there are remaining values.
func (i *LongValuesIterator) HasNext() bool {
	return i.pos < int(i.p.Size())
}

func (i *LongValuesIterator) Next() int64 {
	if i.HasNext() {
		v := i.p.Get(int64(i.pos))
		i.pos++
		return v
	}
	return 0
}

func (p *LongValues) Iterator() *LongValuesIterator {
	iterator := &LongValuesIterator{
		p:   p,
		pos: -1,
	}
	return iterator
}

type LongValuesBuilder struct {
	pageShift, pageMask     int
	acceptableOverheadRatio float64
	pending                 []int64
	size                    int64
	values                  []Reader
	valuesOff               int
	pendingOff              int
}

const (
	MIN_PAGE_SIZE      = 64
	MAX_PAGE_SIZE      = 1 << 20
	INITIAL_PAGE_COUNT = 16
)

func NewLongValuesBuilder(pageSize int, acceptableOverheadRatio float64) *LongValuesBuilder {
	pageShift := checkBlockSize(pageSize, MIN_PAGE_SIZE, MAX_PAGE_SIZE)
	pageMask := pageSize - 1

	return &LongValuesBuilder{
		pageShift:               pageShift,
		pageMask:                pageMask,
		acceptableOverheadRatio: acceptableOverheadRatio,
		values:                  make([]Reader, INITIAL_PAGE_COUNT),
		pending:                 make([]int64, pageSize),
	}
}

func NewLongValuesBuilderV1() *LongValuesBuilder {
	return &LongValuesBuilder{pending: make([]int64, 0)}
}

// Add a new element to this builder.
func (p *LongValuesBuilder) Add(value int64) {

	p.pending = append(p.pending, value)
}

func (p *LongValuesBuilder) pack() {
	p.packV1(p.pending, p.pendingOff, p.valuesOff, p.acceptableOverheadRatio)
	p.valuesOff += 1
	// reset pending buffer
	p.pending = p.pending[:0]
}

func (p *LongValuesBuilder) packV1(values []int64, numValues, block int, acceptableOverheadRatio float64) error {
	// compute max delta
	minValue := values[0]
	maxValue := values[0]

	for _, value := range values {
		minValue = min(minValue, value)
		maxValue = max(maxValue, value)
	}

	// build a new packed reader
	if minValue == 0 && maxValue == 0 {
		p.values[block] = NewNullReader(numValues)
		return nil
	}

	var err error
	var bitsRequired int
	if minValue < 0 {
		bitsRequired = 64
	} else {
		bitsRequired, err = BitsRequired(maxValue)
		if err != nil {
			return err
		}
	}

	copyBuffer := make([]uint64, len(values))

	m := GetMutable(numValues, bitsRequired, acceptableOverheadRatio)
	for i := 0; i < numValues; {
		size := CloneI64ToU64(values[i:numValues-i], copyBuffer)
		i += m.SetBulk(i, copyBuffer[:size])
	}
	p.values[block] = m
	return nil
}

func CloneI64ToU64(src []int64, dest []uint64) int {
	size := len(src)
	for i := 0; i < size; i++ {
		dest[i] = uint64(src[i])
	}
	return size
}
