package memory

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
)

type Info struct {
	fieldInfo *index.FieldInfo
	norm      int64

	// TODO
	// Term strings and their positions for this field: Map <String termText, ArrayIntList positions>
	// private BytesRefHash terms;
	terms *util.BytesRefHash
	// private SliceByteStartArray sliceArray;
	sliceArray *SliceByteStartArray

	// Terms sorted ascending by term text; computed on demand
	sortedTerms []int

	// Number of added tokens for this field
	numTokens int

	// Number of overlapping tokens for this field
	numOverlapTokens int

	sumTotalTermFreq int64

	maxTermFrequency int

	// the last position encountered in this field for multi field support
	lastPosition int

	// the last offset encountered in this field for multi field support
	lastOffset int

	binaryProducer  *BinaryDocValuesProducer
	numericProducer *NumericDocValuesProducer

	preparedDocValuesAndPointValues bool

	pointValues [][]byte

	minPackedValue   []byte
	maxPackedValue   []byte
	pointValuesCount int
}

func NewInfo(fieldInfo *index.FieldInfo, byteBlockPool *util.ByteBlockPool) *Info {
	sliceArray := NewSliceByteStartArray(util.DEFAULT_CAPACITY)

	info := Info{
		fieldInfo:       fieldInfo,
		terms:           util.NewBytesRefHashV1(byteBlockPool, util.DEFAULT_CAPACITY, sliceArray),
		sliceArray:      sliceArray,
		sortedTerms:     make([]int, 0),
		binaryProducer:  NewBinaryDocValuesProducer(),
		numericProducer: NewNumericDocValuesProducer(),
		pointValues:     make([][]byte, 0),
		minPackedValue:  make([]byte, 0),
		maxPackedValue:  make([]byte, 0),
	}

	return &info
}

func (r *Info) freeze() {

}

// Sorts hashed Terms into ascending order, reusing memory along the way. Note that sorting is lazily
// delayed until required (often it's not required at all). If a sorted view is required then
// hashing + sort + binary search is still faster and smaller than TreeMap usage (which would be an
// alternative and somewhat more elegant approach, apart from more sophisticated Tries / prefix trees).
func (r *Info) sortTerms() {
	if len(r.sortedTerms) == 0 {
		r.sortedTerms = r.terms.Sort()
	}
}

func (r *Info) prepareDocValuesAndPointValues() {

}

func (r *Info) getNormDocValues() index.NumericDocValues {
	return nil
}
