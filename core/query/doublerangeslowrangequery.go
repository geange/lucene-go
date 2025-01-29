package query

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

type DoubleRangeSlowRangeQuery struct {
	*BinaryRangeFieldRangeQuery

	field string
	mins  []float64
	maxs  []float64
}

func NewDoubleRangeSlowRangeQuery(field string, minNums, maxNums []float64, queryType QueryType) (*DoubleRangeSlowRangeQuery, error) {
	packedValues, err := encodeDoubleRanges(minNums, maxNums)
	if err != nil {
		return nil, err
	}

	rangeQuery := NewBinaryRangeFieldRangeQuery(field, packedValues, document.DOUBLE_BYTES, len(minNums), queryType)

	return &DoubleRangeSlowRangeQuery{
		BinaryRangeFieldRangeQuery: rangeQuery,
		mins:                       minNums,
		field:                      field,
		maxs:                       maxNums,
	}, nil
}

func (q *DoubleRangeSlowRangeQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	return q.createWeight(q, scoreMode, boost), nil
}

func (q *DoubleRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return q, nil
}

func (q *DoubleRangeSlowRangeQuery) Visit(visitor index.QueryVisitor) error {
	return rangeQueryVisit(q.field, q, visitor)
}

func encodeDoubleRanges(mins, maxs []float64) ([]byte, error) {
	dst := make([]byte, 2*document.DOUBLE_BYTES*len(mins))
	if err := verifyAndEncodeFloat64(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
