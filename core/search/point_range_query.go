package search

import (
	"bytes"
	"errors"

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

func NewPointRangeQuery(field string, numDims int, lowerPoint []byte, upperPoint []byte,
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
	weight := &pointRangeQueryConstantScoreWeight{
		p: p,
	}
	weight.ConstantScoreWeight = NewConstantScoreWeight(boost, p, weight)
	return weight, nil
}

type pointRangeQueryConstantScoreWeight struct {
	*ConstantScoreWeight
	p *PointRangeQuery
}

func (r *pointRangeQueryConstantScoreWeight) matches(packedValue []byte) bool {
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

func (r *pointRangeQueryConstantScoreWeight) relate(minPackedValue, maxPackedValue []byte) index.Relation {
	crosses := false

	for dim := 0; dim < r.p.numDims; dim++ {
		offset := dim * r.p.bytesPerDim

		toIndex := offset + r.p.bytesPerDim

		if bytes.Compare(minPackedValue[offset:toIndex], r.p.upperPoint[offset:toIndex]) > 0 ||
			bytes.Compare(maxPackedValue[offset:toIndex], r.p.lowerPoint[offset:toIndex]) < 0 {
			return index.CELL_OUTSIDE_QUERY
		}

		crosses = crosses || (bytes.Compare(minPackedValue[offset:toIndex], r.p.lowerPoint[offset:toIndex]) < 0 ||
			bytes.Compare(maxPackedValue[offset:toIndex], r.p.upperPoint[offset:toIndex]) > 0)
	}

	if crosses {
		return index.CELL_CROSSES_QUERY
	}
	return index.CELL_INSIDE_QUERY
}

func (r *pointRangeQueryConstantScoreWeight) getIntersectVisitor(result DocIdSetBuilder) index.IntersectVisitor {
	panic("")
}

func (r *pointRangeQueryConstantScoreWeight) ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error) {
	panic("")
}

func (r *pointRangeQueryConstantScoreWeight) Scorer(ctx *index.LeafReaderContext) (Scorer, error) {
	panic("")
}

func (r *pointRangeQueryConstantScoreWeight) IsCacheable(ctx *index.LeafReaderContext) bool {
	return true
}

func (p *PointRangeQuery) Rewrite(reader index.IndexReader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PointRangeQuery) Visit(visitor QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}

func copyOfSubArray(bs []byte, from, to int) []byte {
	newBytes := make([]byte, to-from)
	copy(newBytes, bs[from:])
	return newBytes
}
