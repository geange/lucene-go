package query

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

type LongRangeSlowRangeQuery struct {
	*BinaryRangeFieldRangeQuery

	field string
	mins  []int64
	maxs  []int64
}

func NewLongRangeSlowRangeQuery(field string, mins, maxs []int64, queryType QueryType) (*LongRangeSlowRangeQuery, error) {
	packedValues, err := encodeLongRanges(mins, maxs)
	if err != nil {
		return nil, err
	}

	rangeQuery := NewBinaryRangeFieldRangeQuery(field, packedValues, document.INTEGER_BYTES, len(mins), queryType)

	return &LongRangeSlowRangeQuery{
		BinaryRangeFieldRangeQuery: rangeQuery,
		mins:                       mins,
		field:                      field,
		maxs:                       maxs,
	}, nil
}

func (q *LongRangeSlowRangeQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	return q.createWeight(q, scoreMode, boost), nil
}

func (q *LongRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return q, nil
}

func (q *LongRangeSlowRangeQuery) Visit(visitor index.QueryVisitor) error {
	return rangeQueryVisit(q.field, q, visitor)
}

func encodeLongRanges(mins, maxs []int64) ([]byte, error) {
	dst := make([]byte, 2*document.LONG_BYTES*len(mins))
	if err := verifyAndEncodeInt64(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
