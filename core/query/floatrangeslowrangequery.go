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

func (q *FloatRangeSlowRangeQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	return q.createWeight(q, scoreMode, boost), nil
}

func (q *FloatRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return q, nil
}

func (q *FloatRangeSlowRangeQuery) Visit(visitor index.QueryVisitor) error {
	return rangeQueryVisit(q.field, q, visitor)
}

func encodeFloatRanges(mins, maxs []float32) ([]byte, error) {
	dst := make([]byte, 2*document.FLOAT_BYTES*len(mins))
	if err := verifyAndEncodeFloat32(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
