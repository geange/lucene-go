package query

import (
	"context"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/types"
)

type BinaryRangeFieldRangeQuery struct {
	field                string
	queryPackedValue     []byte
	numBytesPerDimension int
	numDims              int
	queryType            QueryType
}

func (b *BinaryRangeFieldRangeQuery) createWeight(query search.Query, scoreMode search.ScoreMode, boost float64) search.Weight {
	weight := &binaryRangeFieldRangeWeight{
		query:     b,
		scoreMode: scoreMode,
	}
	weight.ConstantScoreWeight = search.NewConstantScoreWeight(boost, query, weight)
	return weight
}

func (b *BinaryRangeFieldRangeQuery) getValues(reader index.LeafReader, field string) (*BinaryRangeDocValues, error) {
	binaryDocValues, err := reader.GetBinaryDocValues(field)
	if err != nil {
		return nil, err
	}
	return NewBinaryRangeDocValues(binaryDocValues, b.numDims, b.numBytesPerDimension), nil
}

var _ search.Weight = &binaryRangeFieldRangeWeight{}

type binaryRangeFieldRangeWeight struct {
	*search.ConstantScoreWeight

	query     *BinaryRangeFieldRangeQuery
	scoreMode search.ScoreMode
}

func (b *binaryRangeFieldRangeWeight) Scorer(ctx index.LeafReaderContext) (search.Scorer, error) {

	values, err := b.query.getValues(ctx.LeafReader(), b.query.field)
	if err != nil {
		return nil, err
	}

	if values == nil {
		return nil, nil
	}

	iterator := &binaryRangeFieldRangeWeightTwoPhaseIterator{
		weight: b,
		values: values,
	}

	return search.NewConstantScoreScorer(b, b.Score(), b.scoreMode, search.AsDocIdSetIterator(iterator))
}

var _ search.TwoPhaseIterator = &binaryRangeFieldRangeWeightTwoPhaseIterator{}

type binaryRangeFieldRangeWeightTwoPhaseIterator struct {
	weight *binaryRangeFieldRangeWeight
	values *BinaryRangeDocValues
}

func (b *binaryRangeFieldRangeWeightTwoPhaseIterator) Approximation() types.DocIdSetIterator {
	return b.values
}

func (b *binaryRangeFieldRangeWeightTwoPhaseIterator) Matches() (bool, error) {
	query := b.weight.query

	return matches(context.TODO(),
		query.queryType,
		query.queryPackedValue,
		b.values.getPackedValue(),
		query.numDims,
		query.numBytesPerDimension,
	), nil
}

func (b *binaryRangeFieldRangeWeightTwoPhaseIterator) MatchCost() float64 {
	return float64(len(b.weight.query.queryPackedValue))

}

func (b *binaryRangeFieldRangeWeight) IsCacheable(ctx index.LeafReaderContext) bool {
	return index.IsCacheable(ctx, b.query.field)
}

func rangeQueryVisit(field string, query search.Query, visitor search.QueryVisitor) error {
	if visitor.AcceptField(field) {
		return visitor.VisitLeaf(query)
	}
	return nil
}

func (b *BinaryRangeFieldRangeQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func NewBinaryRangeFieldRangeQuery(field string, queryPackedValue []byte, numBytesPerDimension int,
	numDims int, queryType QueryType) *BinaryRangeFieldRangeQuery {
	return &BinaryRangeFieldRangeQuery{
		field:                field,
		queryPackedValue:     queryPackedValue,
		numBytesPerDimension: numBytesPerDimension,
		numDims:              numDims,
		queryType:            queryType,
	}
}
