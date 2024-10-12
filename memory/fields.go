package memory

import (
	"bytes"
	"slices"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/version"
)

type Fields struct {
	fields *treemap.Map[string, *info]

	idx *Index
}

func (r *Index) newFields(kv *treemap.Map[string, *info]) *Fields {
	return &Fields{fields: kv, idx: r}
}

func (m *Fields) Names() []string {
	return m.fields.Keys()
}

func (m *Fields) Terms(field string) (index.Terms, error) {
	info, ok := m.fields.Get(field)
	if !ok {
		return nil, nil
	}

	if info.numTokens <= 0 {
		return nil, nil
	}

	return m.idx.newTerms(info), nil
}

func (m *Fields) Size() int {
	return m.fields.Size()
}

// info: Index data structure for a field;
// contains the tokenized term texts and their positions.
type info struct {
	index     *Index
	fieldInfo *document.FieldInfo
	norm      *int64

	// Term strings and their positions for this field: map<termText:string, positions:[]int>
	terms      *bytesref.BytesHash
	sliceArray *sliceByteStartArray

	sortedTerms []int // terms sorted ascending by term text; computed on demand
	numTokens   int   // Number of added tokens for this field

	// Number of overlapping tokens for this field
	numOverlapTokens int
	sumTotalTermFreq int64
	maxTermFrequency int

	lastPosition int // the last position encountered in this field for multi field support
	lastOffset   int // the last offset encountered in this field for multi field support

	binaryProducer                  *binaryDocValuesProducer
	numericProducer                 *numericDocValuesProducer
	preparedDocValuesAndPointValues bool
	pointValues                     [][]byte
	minPackedValue                  []byte
	maxPackedValue                  []byte
	pointValuesCount                int
}

func (r *info) freeze() {
	r.sortTerms()
	r.prepareDocValuesAndPointValues()
	r.getNormDocValues()
}

// Sorts hashed Terms into ascending order, reusing memory along the way. Note that sorting is lazily
// delayed until required (often it's not required at all). If a sorted view is required then
// hashing + sort + binary search is still faster and smaller than TreeMap usage (which would be an
// alternative and somewhat more elegant approach, apart from more sophisticated Tries / prefix trees).
func (r *info) sortTerms() {
	if len(r.sortedTerms) == 0 {
		r.sortedTerms = r.terms.Sort()
	}
}

func (r *info) prepareDocValuesAndPointValues() {
	if r.preparedDocValuesAndPointValues {
		return
	}

	dvType := r.fieldInfo.GetDocValuesType()
	switch dvType {
	case document.DOC_VALUES_TYPE_NUMERIC, document.DOC_VALUES_TYPE_SORTED_NUMERIC:
		r.numericProducer.prepareForUsage()
	case document.DOC_VALUES_TYPE_BINARY, document.DOC_VALUES_TYPE_SORTED, document.DOC_VALUES_TYPE_SORTED_SET:
		r.binaryProducer.prepareForUsage()
	}

	if r.pointValues != nil {
		numDimensions := r.fieldInfo.GetPointDimensionCount()
		numBytesPerDimension := r.fieldInfo.GetPointNumBytes()

		if numDimensions == 1 {
			// PointInSetQuery.MergePointVisitor expects values to be visited in increasing order,
			// this is a 1d optimization which has to be done here too. Otherwise we emit values
			// out of order which causes mismatches.
			slices.SortFunc(r.pointValues[:r.pointValuesCount], bytes.Compare)
			r.minPackedValue = slices.Clone(r.pointValues[0])
			r.maxPackedValue = slices.Clone(r.pointValues[r.pointValuesCount-1])
			return
		}

		r.minPackedValue = slices.Clone(r.pointValues[0])
		r.maxPackedValue = slices.Clone(r.pointValues[0])
		for i := 0; i < r.pointValuesCount; i++ {
			pointValue := r.pointValues[i]
			for dim := 0; dim < numDimensions; dim++ {

				fromIndex := dim * numBytesPerDimension
				toIndex := fromIndex + numBytesPerDimension

				dimValues := pointValue[fromIndex:toIndex]

				if bytes.Compare(dimValues, r.minPackedValue[fromIndex:toIndex]) < 0 {
					copy(r.minPackedValue[fromIndex:toIndex], dimValues)
				}

				if bytes.Compare(dimValues, r.maxPackedValue[fromIndex:toIndex]) > 0 {
					copy(r.maxPackedValue[fromIndex:toIndex], dimValues)
				}
			}
		}

		return
	}
	r.preparedDocValuesAndPointValues = true
}

func (r *info) getNormDocValues() index.NumericDocValues {
	if r.norm == nil {
		invertState := index.NewFieldInvertState(
			int(version.Last.Major()),
			r.fieldInfo.Name(),
			r.fieldInfo.GetIndexOptions(),
			r.lastPosition,
			r.numTokens,
			r.numOverlapTokens,
			0,
			r.maxTermFrequency,
			r.terms.Size())

		value := r.index.normSimilarity.ComputeNorm(invertState)
		r.norm = &value
	}
	return newNumericDocValues(*r.norm)

}
