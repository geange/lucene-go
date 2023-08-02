package query

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.Query = &IntRangeSlowRangeQuery{}

type IntRangeSlowRangeQuery struct {
	*BinaryRangeFieldRangeQuery

	field string
	mins  []int32
	maxs  []int32
}

func NewIntRangeSlowRangeQuery(field string, mins, maxs []int32, queryType QueryType) (*IntRangeSlowRangeQuery, error) {
	packedValues, err := encodeIntRanges(mins, maxs)
	if err != nil {
		return nil, err
	}

	rangeQuery := NewBinaryRangeFieldRangeQuery(field, packedValues, document.INTEGER_BYTES, len(mins), queryType)

	return &IntRangeSlowRangeQuery{
		BinaryRangeFieldRangeQuery: rangeQuery,
		mins:                       mins,
		field:                      field,
		maxs:                       maxs,
	}, nil
}

func (i *IntRangeSlowRangeQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	return i.createWeight(i, scoreMode, boost), nil
}

func (i *IntRangeSlowRangeQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return i, nil
}

func (i *IntRangeSlowRangeQuery) Visit(visitor index.QueryVisitor) error {
	return rangeQueryVisit(i.field, i, visitor)
}

func encodeIntRanges(mins, maxs []int32) ([]byte, error) {
	dst := make([]byte, 2*document.INTEGER_BYTES*len(mins))
	if err := verifyAndEncodeInt32(mins, maxs, dst); err != nil {
		return nil, err
	}
	return dst, nil
}
