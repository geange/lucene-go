package memory

import (
	"github.com/geange/lucene-go/core/util"
	"sort"
)

type BinaryDocValuesProducer struct {
	dvBytesValuesSet *util.BytesRefHash
	bytesIds         []int
}

func NewBinaryDocValuesProducer() *BinaryDocValuesProducer {
	return &BinaryDocValuesProducer{}
}

func (r *BinaryDocValuesProducer) prepareForUsage() {
	r.bytesIds = r.dvBytesValuesSet.Sort()
}

type NumericDocValuesProducer struct {
	dvLongValues []int
	count        int
}

func NewNumericDocValuesProducer() *NumericDocValuesProducer {
	return &NumericDocValuesProducer{}
}

func (r *NumericDocValuesProducer) prepareForUsage() {
	sort.Ints(r.dvLongValues[0:r.count])
}
