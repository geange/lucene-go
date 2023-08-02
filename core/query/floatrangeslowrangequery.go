package query

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

type FloatRangeSlowRangeQuery struct {
	*BinaryRangeFieldRangeQuery

	field string
	mins  []float32
	maxs  []float32
}

func NewFloatRangeSlowRangeQuery(field string, mins, maxs []float32, queryType QueryType) (*FloatRangeSlowRangeQuery, error) {
	packedValues, err := encodeFloatRanges(mins, maxs)
	if err != nil {
		return nil, err
	}

	rangeQuery := NewBinaryRangeFieldRangeQuery(field, packedValues, document.FLOAT_BYTES, len(mins), queryType)

	return &FloatRangeSlowRangeQuery{
		BinaryRangeFieldRangeQuery: rangeQuery,
		mins:                       mins,
		field:                      field,
		maxs:                       maxs,
	}, nil
}

func (i *FloatRangeSlowRangeQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	return i.createWeight(i, scoreMode, boost), nil
}

func (i *FloatRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return i, nil
}

func (i *FloatRangeSlowRangeQuery) Visit(visitor index.QueryVisitor) error {
	return rangeQueryVisit(i.field, i, visitor)
}

func encodeFloatRanges(mins, maxs []float32) ([]byte, error) {
	dst := make([]byte, 2*document.FLOAT_BYTES*len(mins))
	if err := verifyAndEncodeFloat32(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
