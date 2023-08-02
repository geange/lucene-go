package index

import (
	"math"
	"slices"

	"github.com/bits-and-blooms/bitset"
)

// FieldUpdatesBuffer
// This class efficiently buffers numeric and binary field updates and stores terms,
// values and metadata in a memory efficient way without creating large amounts of objects. Update
// terms are stored without de-duplicating the update term. In general we try to optimize for several
// use-cases. For instance we try to use constant space for update terms field since the common case always
// updates on the same field. Also for docUpTo we try to optimize for the case when updates should be applied
// to all docs ie. docUpTo=Integer.MAX_VALUE. In other cases each update will likely have a different docUpTo.
// Along the same lines this impl optimizes the case when all updates have a item. Lastly, if all updates
// share the same item for a numeric field we only store the item once.
type FieldUpdatesBuffer struct {
	numUpdates    int
	termValues    [][]byte
	termSortState []int
	byteValues    [][]byte
	docsUpTo      []int
	numericValues []int64
	hasValues     *bitset.BitSet
	maxNumeric    int64
	minNumeric    int64
	fields        []string
	isNumeric     bool
	finished      bool
}

func NewNumberFieldUpdatesBuffer(initialValue *NumericDocValuesUpdate, docUpTo int) *FieldUpdatesBuffer {
	buffer := newFieldUpdatesBuffer(initialValue, docUpTo, true)
	if initialValue.HasValue() {
		num := initialValue.GetValue()
		buffer.numericValues = []int64{num}
		buffer.maxNumeric = num
		buffer.minNumeric = num
	} else {
		buffer.numericValues = []int64{0}
	}
	return buffer
}

func NewBinaryFieldUpdatesBuffer(initialValue *BinaryDocValuesUpdate, docUpTo int) *FieldUpdatesBuffer {
	buffer := newFieldUpdatesBuffer(initialValue, docUpTo, false)
	if initialValue.HasValue() {
		buffer.byteValues = append(buffer.byteValues, initialValue.GetValue())
	}
	return buffer
}

func newFieldUpdatesBuffer(initialValue DocValuesUpdate, docUpTo int, isNumeric bool) *FieldUpdatesBuffer {

	buffer := &FieldUpdatesBuffer{
		numUpdates:    1,
		termValues:    [][]byte{initialValue.GetTerm().Bytes()},
		termSortState: nil,
		byteValues:    nil,
		docsUpTo:      []int{docUpTo},
		numericValues: nil,
		hasValues:     nil,
		maxNumeric:    math.MaxInt64,
		minNumeric:    math.MinInt64,
		fields:        []string{initialValue.GetTerm().Field()},
		isNumeric:     isNumeric,
		finished:      false,
	}

	if !initialValue.HasValue() {
		buffer.hasValues = bitset.New(1)
	}

	return buffer
}

func (f *FieldUpdatesBuffer) getMaxNumeric() int64 {
	if f.minNumeric == math.MaxInt64 && f.maxNumeric == math.MinInt64 {
		return 0
	}
	return f.maxNumeric
}

func (f *FieldUpdatesBuffer) getMinNumeric() int64 {
	if f.minNumeric == math.MaxInt64 && f.maxNumeric == math.MinInt64 {
		return 0
	}
	return f.minNumeric
}

func (f *FieldUpdatesBuffer) add(field string, docUpTo, ord int, hasValue bool) error {
	if f.fields[0] != (field) || len(f.fields) != 1 {
		if len(f.fields) <= ord {
			f.fields = slices.Grow(f.fields, ord+1)
		}
		f.fields[ord] = field
	}

	if f.docsUpTo[0] != docUpTo || len(f.docsUpTo) != 1 {
		if len(f.docsUpTo) <= ord {
			f.docsUpTo = slices.Grow(f.docsUpTo, ord+1)
		}
		f.docsUpTo[ord] = docUpTo
	}

	if f.hasValues == nil {
		f.hasValues = bitset.New(uint(ord + 1))
	}
	if hasValue {
		f.hasValues.Set(uint(ord))
	}
	return nil
}

func (f *FieldUpdatesBuffer) addUpdateInt(term Term, value int64, docUpTo int) error {
	ord := f.append(term)
	field := term.Field()
	if err := f.add(field, docUpTo, ord, true); err != nil {
		return err
	}
	f.minNumeric = min(f.minNumeric, value)
	f.maxNumeric = max(f.maxNumeric, value)

	if f.numericValues[0] != value || len(f.numericValues) != 1 {
		if len(f.numericValues) <= ord {
			f.numericValues = slices.Grow(f.numericValues, ord+1)
		}
		f.numericValues[ord] = value
	}
	return nil
}

func (f *FieldUpdatesBuffer) addNoValue(term Term, docUpTo int) error {
	ord := f.append(term)
	return f.add(term.Field(), docUpTo, ord, false)
}

func (f *FieldUpdatesBuffer) addUpdateBytes(term Term, value []byte, docUpTo int) error {
	ord := f.append(term)
	f.byteValues = append(f.byteValues, value)
	return f.add(term.Field(), docUpTo, ord, true)
}

func (f *FieldUpdatesBuffer) append(term Term) int {
	f.termValues = append(f.termValues, term.Bytes())
	numUpdates := f.numUpdates
	f.numUpdates++
	return numUpdates
}

func (f *FieldUpdatesBuffer) IsNumeric() bool {
	return f.isNumeric
}

func (f *FieldUpdatesBuffer) hasSingleValue() bool {
	// we only do this optimization for numerics so far.
	return f.isNumeric && len(f.numericValues) == 1
}

func (f *FieldUpdatesBuffer) getNumericValue(idx int) int64 {
	if f.hasValues != nil && f.hasValues.Test(uint(idx)) == false {
		return 0
	}
	return f.numericValues[min(len(f.numericValues), idx)]
}

func (f *FieldUpdatesBuffer) Range() {

}
