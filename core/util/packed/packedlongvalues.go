package packed

import (
	"io"
	"slices"

	"github.com/geange/lucene-go/core/types"
)

var _ types.LongValues = &PackedLongValues{}

// PackedLongValues compress integers into a PackedLongValues instance.
type PackedLongValues struct {
	values    []Reader
	pageShift int
	pageMask  uint64
	size      int
}

func NewPackedLongValues(values []Reader, pageShift int, pageMask uint64, size int) *PackedLongValues {
	return &PackedLongValues{values: values, pageShift: pageShift, pageMask: pageMask, size: size}
}

func (p *PackedLongValues) Size() int {
	return p.size
}

func (p *PackedLongValues) Get(index int) (uint64, error) {
	return p.getWithFnGet(index, p.get)
}

func (p *PackedLongValues) getWithFnGet(index int, fnGet func(block int, element int) (uint64, error)) (uint64, error) {
	block := index >> p.pageShift
	element := int(uint64(index) & p.pageMask)
	return fnGet(block, element)
}

func (p *PackedLongValues) get(block int, element int) (uint64, error) {
	return p.values[block].Get(element)
}

type LongValuesIteratorSPI interface {
	Get(index int) (uint64, error)
	Size() int
}

type PackedLongValuesIterator interface {
	HasNext() bool
	Next() (uint64, error)
}

type packedLongValuesIterator struct {
	spi LongValuesIteratorSPI
	pos int
}

// HasNext
// Whether or not there are remaining values.
func (i *packedLongValuesIterator) HasNext() bool {
	return i.pos < i.spi.Size()
}

func (i *packedLongValuesIterator) Next() (uint64, error) {
	if i.HasNext() {
		v, err := i.spi.Get(i.pos)
		if err != nil {
			return 0, err
		}
		i.pos++
		return v, nil
	}
	return 0, io.EOF
}

func (p *PackedLongValues) Iterator() PackedLongValuesIterator {
	return p.iteratorWithSPI(p)
}

func (p *PackedLongValues) iteratorWithSPI(spi LongValuesIteratorSPI) *packedLongValuesIterator {
	iterator := &packedLongValuesIterator{
		spi: spi,
		pos: 0,
	}
	return iterator
}

type PackedLongValuesBuilder struct {
	pageShift               int
	pageMask                uint64
	acceptableOverheadRatio float64
	pending                 []int64
	pageSize                int
	size                    int
	values                  []Reader
}

const (
	MIN_PAGE_SIZE      = 64
	MAX_PAGE_SIZE      = 1 << 20
	INITIAL_PAGE_COUNT = 16
)

func NewPackedLongValuesBuilder(pageSize int, acceptableOverheadRatio float64) *PackedLongValuesBuilder {
	pageShift := checkBlockSize(pageSize, MIN_PAGE_SIZE, MAX_PAGE_SIZE)
	pageMask := uint64(pageSize) - 1

	return &PackedLongValuesBuilder{
		pageShift:               pageShift,
		pageMask:                pageMask,
		acceptableOverheadRatio: acceptableOverheadRatio,
		values:                  make([]Reader, 0, INITIAL_PAGE_COUNT),
		pending:                 make([]int64, 0, pageSize),
		pageSize:                pageSize,
	}
}

func (p *PackedLongValuesBuilder) Build() (*PackedLongValues, error) {
	if err := p.finish(); err != nil {
		return nil, err
	}
	p.pending = nil
	values := slices.Clone(p.values)
	longValues := NewPackedLongValues(values, p.pageShift, p.pageMask, p.size)
	return longValues, nil
}

func (p *PackedLongValuesBuilder) Size() int {
	return p.size
}

// Add a new element to this builder.
func (p *PackedLongValuesBuilder) Add(value int64) error {
	return p.addWithFnPack(value, p.pack)
}

func (p *PackedLongValuesBuilder) addWithFnPack(value int64, pack func() error) error {
	if len(p.pending) == p.pageSize {
		if err := pack(); err != nil {
			return err
		}
	}
	p.pending = append(p.pending, value)
	p.size++
	return nil
}

func (p *PackedLongValuesBuilder) finish() error {
	return p.finishWithFnPack(p.pack)
}

func (p *PackedLongValuesBuilder) finishWithFnPack(pack func() error) error {
	if len(p.pending) > 0 {
		return pack()
	}
	return nil
}

func (p *PackedLongValuesBuilder) pack() error {
	return p.packWithFnPackValues(p.packValues)
}

func (p *PackedLongValuesBuilder) packWithFnPackValues(packValues FnPackValues) error {
	if err := packValues(p.pending, len(p.pending), p.acceptableOverheadRatio); err != nil {
		return err
	}

	// reset pending buffer
	p.pending = p.pending[:0]
	return nil
}

type FnPackValues func(values []int64, numValues int, acceptableOverheadRatio float64) error

func (p *PackedLongValuesBuilder) packValues(values []int64, numValues int, acceptableOverheadRatio float64) error {
	// compute max delta
	minValue := values[0]
	maxValue := values[0]

	for _, value := range values[1:] {
		minValue = min(minValue, value)
		maxValue = max(maxValue, value)
	}

	// build a new packed reader
	if minValue == 0 && maxValue == 0 {
		p.values = append(p.values, NewNullReader(numValues))
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

	m := DefaultGetMutable(numValues, bitsRequired, acceptableOverheadRatio)
	for i := 0; i < numValues; {
		size := CloneI64ToU64(values[i:numValues], copyBuffer)
		i += m.SetBulk(i, copyBuffer[:size])
	}
	p.values = append(p.values, m)
	return nil
}

func CloneI64ToU64(src []int64, dest []uint64) int {
	size := len(src)
	for i := 0; i < size; i++ {
		dest[i] = uint64(src[i])
	}
	return size
}
