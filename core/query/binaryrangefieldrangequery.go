package query

import (
	"context"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
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

func (q *BinaryRangeFieldRangeQuery) createWeight(query index.Query,
	scoreMode index.ScoreMode, boost float64) index.Weight {
	weight := &binaryRangeFieldRangeWeight{
		query:     q,
		scoreMode: scoreMode,
	}
	weight.ConstantScoreWeight = search.NewConstantScoreWeight(boost, query, weight)
	return weight
}

func (q *BinaryRangeFieldRangeQuery) getValues(reader index.LeafReader, field string) (*BinaryRangeDocValues, error) {
	binaryDocValues, err := reader.GetBinaryDocValues(field)
	if err != nil {
		return nil, err
	}
	return NewBinaryRangeDocValues(binaryDocValues, q.numDims, q.numBytesPerDimension), nil
}

var _ index.Weight = &binaryRangeFieldRangeWeight{}

type binaryRangeFieldRangeWeight struct {
	*search.ConstantScoreWeight

	query     *BinaryRangeFieldRangeQuery
	scoreMode index.ScoreMode
}

func (w *binaryRangeFieldRangeWeight) Scorer(ctx index.LeafReaderContext) (index.Scorer, error) {

	values, err := w.query.getValues(ctx.LeafReader(), w.query.field)
	if err != nil {
		return nil, err
	}

	if values == nil {
		return nil, nil
	}

	iterator := &binaryRangeFieldRangeWeightTwoPhaseIterator{
		weight: w,
		values: values,
	}

	return search.NewConstantScoreScorer(w, w.Score(), w.scoreMode, search.AsDocIdSetIterator(iterator))
}

var _ index.TwoPhaseIterator = &binaryRangeFieldRangeWeightTwoPhaseIterator{}

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

func (w *binaryRangeFieldRangeWeight) IsCacheable(ctx index.LeafReaderContext) bool {
	return coreIndex.IsCacheable(ctx, w.query.field)
}

func rangeQueryVisit(field string, query index.Query, visitor index.QueryVisitor) error {
	if visitor.AcceptField(field) {
		return visitor.VisitLeaf(query)
	}
	return nil
}

func (q *BinaryRangeFieldRangeQuery) String(field string) string {
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
