package query

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
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

func (i *LongRangeSlowRangeQuery) CreateWeight(searcher search.IndexSearcher, scoreMode search.ScoreMode, boost float64) (search.Weight, error) {
	return i.createWeight(i, scoreMode, boost), nil
}

func (i *LongRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (search.Query, error) {
	return i, nil
}

func (i *LongRangeSlowRangeQuery) Visit(visitor search.QueryVisitor) error {
	return rangeQueryVisit(i.field, i, visitor)
}

func encodeLongRanges(mins, maxs []int64) ([]byte, error) {
	dst := make([]byte, 2*document.LONG_BYTES*len(mins))
	if err := verifyAndEncodeInt64(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
