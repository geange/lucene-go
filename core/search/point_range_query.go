package search

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/types"
	"io"
	"math"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/index"
)

var _ Query = &PointRangeQuery{}

type PointRangeQuery struct {
	field       string
	numDims     int
	bytesPerDim int
	lowerPoint  []byte
	upperPoint  []byte
	fn          func(dimension int, value []byte) string
}

func NewPointRangeQuery(field string, lowerPoint []byte, upperPoint []byte, numDims int,
	fn func(dimension int, value []byte) string) (*PointRangeQuery, error) {

	if numDims <= 0 {
		return nil, errors.New("numDims must be positive")
	}

	if len(lowerPoint) == 0 {
		return nil, errors.New("lowerPoint has length of zero")
	}

	if len(lowerPoint)%numDims != 0 {
		return nil, errors.New("lowerPoint is not a fixed multiple of numDims")
	}

	if len(lowerPoint) != len(upperPoint) {
		return nil, errors.New("lowerPoint's length not equal upperPoint's length")
	}

	return &PointRangeQuery{
		field:       field,
		numDims:     numDims,
		bytesPerDim: len(lowerPoint) / numDims,
		lowerPoint:  lowerPoint,
		upperPoint:  upperPoint,
		fn:          fn,
	}, nil
}

func (p *PointRangeQuery) String(field string) string {
	sb := new(bytes.Buffer)

	if p.field != field {
		sb.WriteString(p.field)
		sb.WriteString(":")
	}

	// print ourselves as "range per dimension"
	for i := 0; i < p.numDims; i++ {
		if i > 0 {
			sb.WriteString(".")
		}

		startOffset := p.bytesPerDim * i
		sb.WriteString("[")
		sb.WriteString(p.fn(i, copyOfSubArray(p.lowerPoint, startOffset, startOffset+p.bytesPerDim)))
		sb.WriteString(" TO ")
		sb.WriteString(p.fn(i, copyOfSubArray(p.upperPoint, startOffset, startOffset+p.bytesPerDim)))
		sb.WriteString("]")
	}
	return sb.String()
}

func (p *PointRangeQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {

	// We don't use RandomAccessWeight here: it's no good to approximate with "match all docs".
	// This is an inverted structure and should be used in the first pass:
	return p.newPrQueryWeight(boost, scoreMode), nil
}

type prQueryWeight struct {
	*ConstantScoreWeight
	p         *PointRangeQuery
	scoreMode *ScoreMode
}

func (p *PointRangeQuery) newPrQueryWeight(boost float64, scoreMode *ScoreMode) *prQueryWeight {
	weight := &prQueryWeight{
		ConstantScoreWeight: nil,
		p:                   p,
		scoreMode:           scoreMode,
	}
	weight.ConstantScoreWeight = NewConstantScoreWeight(boost, p, weight)
	return weight
}

func (r *prQueryWeight) matches(packedValue []byte) bool {
	for dim := 0; dim < r.p.numDims; dim++ {
		fromIndex := dim * r.p.bytesPerDim
		toIndex := fromIndex + r.p.bytesPerDim

		if bytes.Compare(packedValue[fromIndex:toIndex], r.p.lowerPoint[fromIndex:toIndex]) < 0 {
			// Doc's value is too low, in this dimension
			return false
		}

		if bytes.Compare(packedValue[fromIndex:toIndex], r.p.upperPoint[fromIndex:toIndex]) > 0 {
			// Doc's value is too high, in this dimension
			return false
		}
	}
	return true
}

func (r *prQueryWeight) relate(minPackedValue, maxPackedValue []byte) types.Relation {
	crosses := false

	for dim := 0; dim < r.p.numDims; dim++ {
		offset := dim * r.p.bytesPerDim

		toIndex := offset + r.p.bytesPerDim

		if bytes.Compare(minPackedValue[offset:toIndex], r.p.upperPoint[offset:toIndex]) > 0 ||
			bytes.Compare(maxPackedValue[offset:toIndex], r.p.lowerPoint[offset:toIndex]) < 0 {
			return types.CELL_OUTSIDE_QUERY
		}

		crosses = crosses || (bytes.Compare(minPackedValue[offset:toIndex], r.p.lowerPoint[offset:toIndex]) < 0 ||
			bytes.Compare(maxPackedValue[offset:toIndex], r.p.upperPoint[offset:toIndex]) > 0)
	}

	if crosses {
		return types.CELL_CROSSES_QUERY
	}
	return types.CELL_INSIDE_QUERY
}

func (r *prQueryWeight) getIntersectVisitor(result *DocIdSetBuilder) types.IntersectVisitor {
	return &prQueryVisitor{
		weight: r,
		addr:   nil,
		result: result,
	}
}

var _ types.IntersectVisitor = &prQueryVisitor{}

type prQueryVisitor struct {
	weight *prQueryWeight
	addr   BulkAdder
	result *DocIdSetBuilder
}

func (p *prQueryVisitor) Visit(docID int) error {
	p.addr.Add(docID)
	return nil
}

func (p *prQueryVisitor) VisitLeaf(docID int, packedValue []byte) error {
	if p.weight.matches(packedValue) {
		return p.Visit(docID)
	}
	return nil
}

func (p *prQueryVisitor) VisitIterator(iterator types.DocValuesIterator, packedValue []byte) error {
	if p.weight.matches(packedValue) {
		for {
			doc, err := iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
			err = p.Visit(doc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *prQueryVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	return p.weight.relate(minPackedValue, maxPackedValue)
}

func (p *prQueryVisitor) Grow(count int) {
	p.addr = p.result.Grow(count)
}

func (r *prQueryWeight) getInverseIntersectVisitor(result *bitset.BitSet, cost []int64) types.IntersectVisitor {
	panic("")
}

var _ types.IntersectVisitor = &invPrQueryVisitor{}

type invPrQueryVisitor struct {
	result *bitset.BitSet
	cost   []int64
	weight *prQueryWeight
}

func (r *invPrQueryVisitor) Visit(docID int) error {
	r.result.Clear(uint(docID))
	r.cost[0]--
	return nil
}

func (r *invPrQueryVisitor) VisitLeaf(docID int, packedValue []byte) error {
	if r.weight.matches(packedValue) == false {
		return r.Visit(docID)
	}
	return nil
}

func (r *invPrQueryVisitor) VisitIterator(iterator types.DocValuesIterator, packedValue []byte) error {
	if r.weight.matches(packedValue) == false {
		for {
			doc, err := iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
			err = r.Visit(doc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *invPrQueryVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	relation := r.weight.relate(minPackedValue, maxPackedValue)
	switch relation {
	case types.CELL_INSIDE_QUERY:
		// all points match, skip this subtree
		return types.CELL_OUTSIDE_QUERY
	case types.CELL_OUTSIDE_QUERY:
		// none of the points match, clear all documents
		return types.CELL_INSIDE_QUERY
	default:
		return relation
	}
}

func (r *invPrQueryVisitor) Grow(count int) {
	return
}

func (r *prQueryWeight) ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error) {
	reader, ok := ctx.Reader().(index.LeafReader)
	if !ok {
		return nil, errors.New("get reader error")
	}

	field := r.p.field

	values, exist := reader.GetPointValues(field)
	if !exist {
		return nil, nil
	}

	dimensions, err := values.GetNumIndexDimensions()
	if err != nil {
		return nil, err
	}
	if dimensions != r.p.numDims {
		return nil, fmt.Errorf("field=%s numIndexDimensions not equal", field)
	}

	bytesPerDimension, err := values.GetBytesPerDimension()
	if err != nil {
		return nil, err
	}
	if bytesPerDimension != r.p.bytesPerDim {
		return nil, fmt.Errorf("field=%s bytesPerDim not equal", field)
	}

	allDocsMatch := false
	if values.GetDocCount() == reader.MaxDoc() {
		fieldPackedLower, err := values.GetMinPackedValue()
		if err != nil {
			return nil, err
		}
		fieldPackedUpper, err := values.GetMaxPackedValue()
		if err != nil {
			return nil, err
		}
		allDocsMatch = true

		for i := 0; i < r.p.numDims; i++ {
			offset := i * r.p.bytesPerDim
			toIndex := offset + r.p.bytesPerDim

			if bytes.Compare(r.p.lowerPoint[offset:toIndex], fieldPackedLower[offset:toIndex]) > 0 ||
				bytes.Compare(r.p.upperPoint[offset:toIndex], fieldPackedUpper[offset:toIndex]) < 0 {
				allDocsMatch = false
				break
			}
		}
	}

	if allDocsMatch {
		// all docs have a value and all points are within bounds, so everything matches
		return &allDocsScorerSupplier{
			reader:    reader,
			weight:    r,
			scoreMode: r.scoreMode,
		}, nil
	} else {
		result := NewDocIdSetBuilderV2(reader.MaxDoc(), values, field)
		return &notAllDocsScorerSupplier{
			result:    result,
			visitor:   r.getIntersectVisitor(result),
			reader:    reader,
			values:    values,
			cost:      -1,
			weight:    r,
			scoreMode: r.scoreMode,
		}, nil
	}

}

var _ ScorerSupplier = &allDocsScorerSupplier{}

type allDocsScorerSupplier struct {
	reader    index.LeafReader
	weight    *prQueryWeight
	scoreMode *ScoreMode
}

func (r *allDocsScorerSupplier) Get(leadCost int64) (Scorer, error) {
	return NewConstantScoreScorer(r.weight, r.weight.Score(), r.scoreMode, types.DocIdSetIteratorAll(r.reader.MaxDoc()))
}

func (a *allDocsScorerSupplier) Cost() int64 {
	return int64(a.reader.MaxDoc())
}

var _ ScorerSupplier = &notAllDocsScorerSupplier{}

type notAllDocsScorerSupplier struct {
	result    *DocIdSetBuilder
	visitor   types.IntersectVisitor
	reader    index.LeafReader
	values    types.PointValues
	cost      int64
	weight    *prQueryWeight
	scoreMode *ScoreMode
}

func (r *notAllDocsScorerSupplier) Get(leadCost int64) (Scorer, error) {
	if r.values.GetDocCount() == r.reader.MaxDoc() &&
		r.values.GetDocCount() == r.values.Size() &&
		r.Cost() > int64(r.reader.MaxDoc()/2) {

		// If all docs have exactly one value and the cost is greater
		// than half the leaf size then maybe we can make things faster
		// by computing the set of documents that do NOT match the range
		result := bitset.New(uint(r.reader.MaxDoc()))
		for i := range result.Bytes() {
			result.Bytes()[i] = math.MaxUint64
		}
		cost := []int64{int64(r.reader.MaxDoc())}
		err := r.values.Intersect(nil, r.weight.getInverseIntersectVisitor(result, cost))
		if err != nil {
			return nil, err
		}
		iterator := index.NewBitSetIterator(result, cost[0])
		return NewConstantScoreScorer(r.weight, r.weight.Score(), r.scoreMode, iterator)
	}

	err := r.values.Intersect(nil, r.visitor)
	if err != nil {
		return nil, err
	}
	iterator := r.result.Build().Iterator()
	return NewConstantScoreScorer(r.weight, r.weight.Score(), r.scoreMode, iterator)
}

func (r *notAllDocsScorerSupplier) Cost() int64 {
	if r.cost == -1 {
		// Computing the cost may be expensive, so only do it if necessary
		cost, _ := r.values.EstimateDocCount(r.visitor)
		r.cost = int64(cost)
		//assert cost >= 0;
	}
	return r.cost
}

func (r *prQueryWeight) Scorer(ctx *index.LeafReaderContext) (Scorer, error) {
	scorerSupplier, err := r.ScorerSupplier(ctx)
	if err != nil {
		return nil, err
	}
	if scorerSupplier == nil {
		return nil, nil
	}
	return scorerSupplier.Get(math.MaxInt32)
}

func (r *prQueryWeight) IsCacheable(ctx *index.LeafReaderContext) bool {
	return true
}

func (p *PointRangeQuery) Rewrite(reader index.Reader) (Query, error) {
	return p, nil
}

func (p *PointRangeQuery) Visit(visitor QueryVisitor) (err error) {
	if visitor.AcceptField(p.field) {
		err := visitor.VisitLeaf(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyOfSubArray(bs []byte, from, to int) []byte {
	newBytes := make([]byte, to-from)
	copy(newBytes, bs[from:])
	return newBytes
}
