package memory

import (
	"sort"

	"github.com/geange/lucene-go/core/util/bytesref"
)

type binaryDocValuesProducer struct {
	dvBytesValuesSet *bytesref.BytesHash
	bytesIds         []int
}

func newBinaryDocValuesProducer() *binaryDocValuesProducer {
	return &binaryDocValuesProducer{}
}

func (r *binaryDocValuesProducer) prepareForUsage() {
	r.bytesIds = r.dvBytesValuesSet.Sort()
}

type numericDocValuesProducer struct {
	dvLongValues []int
	count        int
}

func newNumericDocValuesProducer() *numericDocValuesProducer {
	return &numericDocValuesProducer{}
}

func (r *numericDocValuesProducer) prepareForUsage() {
	sort.Ints(r.dvLongValues[0:r.count])
}
