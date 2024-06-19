package query

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
)

type DoubleRangeSlowRangeQuery struct {
	*BinaryRangeFieldRangeQuery

	field string
	mins  []float64
	maxs  []float64
}

func NewDoubleRangeSlowRangeQuery(field string, mins, maxs []float64, queryType QueryType) (*DoubleRangeSlowRangeQuery, error) {
	packedValues, err := encodeDoubleRanges(mins, maxs)
	if err != nil {
		return nil, err
	}

	rangeQuery := NewBinaryRangeFieldRangeQuery(field, packedValues, document.DOUBLE_BYTES, len(mins), queryType)

	return &DoubleRangeSlowRangeQuery{
		BinaryRangeFieldRangeQuery: rangeQuery,
		mins:                       mins,
		field:                      field,
		maxs:                       maxs,
	}, nil
}

func (i *DoubleRangeSlowRangeQuery) CreateWeight(searcher search.IndexSearcher, scoreMode search.ScoreMode, boost float64) (search.Weight, error) {
	return i.createWeight(i, scoreMode, boost), nil
}

func (i *DoubleRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (search.Query, error) {
	return i, nil
}

func (i *DoubleRangeSlowRangeQuery) Visit(visitor search.QueryVisitor) error {
	return rangeQueryVisit(i.field, i, visitor)
}

func encodeDoubleRanges(mins, maxs []float64) ([]byte, error) {
	dst := make([]byte, 2*document.DOUBLE_BYTES*len(mins))
	if err := verifyAndEncodeFloat64(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
