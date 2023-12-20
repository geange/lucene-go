package search

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
)

var _ Query = &PointInSetQuery{}

type PointInSetQuery struct {
	sortedPackedPoints         *index.PrefixCodedTerms
	sortedPackedPointsHashCode int
	field                      string
	numDims                    int
	bytesPerDim                int
}

func NewPointInSetQuery(ctx context.Context, field string, numDims int, bytesPerDim int, packedPoints [][]byte) (*PointInSetQuery, error) {
	builder := index.NewPrefixCodedTermsBuilder()
	for i := 1; i < len(packedPoints); i++ {
		previous := packedPoints[i-1]
		current := packedPoints[i]

		if v := bytes.Compare(previous, current); v == 0 {
			continue // deduplicate
		} else if v > 0 {
			return nil, errors.New("values are out of order")
		}

		if i == 1 {
			err := builder.AddBytes(ctx, field, previous)
			if err != nil {
				return nil, err
			}
		}
		err := builder.AddBytes(ctx, field, current)
		if err != nil {
			return nil, err
		}
	}

	return &PointInSetQuery{
		sortedPackedPoints: builder.Finish(),
		field:              field,
		numDims:            numDims,
		bytesPerDim:        bytesPerDim,
	}, nil
}

func (p *PointInSetQuery) String(field string) string {
	return p.field
}

func (p *PointInSetQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	weight := &pisQueryWeight{
		ConstantScoreWeight: nil,
		p:                   p,
		scoreMode:           scoreMode,
	}

	weight.ConstantScoreWeight = NewConstantScoreWeight(boost, p, weight)
	return weight, nil
}

type pisQueryWeight struct {
	*ConstantScoreWeight
	p         *PointInSetQuery
	scoreMode *ScoreMode
}

func (r *pisQueryWeight) Scorer(ctx *index.LeafReaderContext) (Scorer, error) {
	reader := ctx.Reader().(index.LeafReader)
	values, ok := reader.GetPointValues(r.p.field)
	if !ok {
		// No docs in this segment/field indexed any points
		return nil, nil
	}

	if dims, _ := values.GetNumIndexDimensions(); dims != r.p.numDims {
		return nil, errors.New("numIndexDims not fit")
	}

	if dims, _ := values.GetBytesPerDimension(); dims != r.p.bytesPerDim {
		return nil, errors.New("bytesPerDim not fit")
	}

	result := NewDocIdSetBuilderV2(reader.MaxDoc(), values, r.p.field)
	if r.p.numDims == 1 {
		// We optimize this common case, effectively doing a merge sort of the indexed values vs the queried set:
		visitor, err := r.p.NewMergePointVisitor(r.p.sortedPackedPoints, result)
		if err != nil {
			return nil, err
		}
		err = values.Intersect(nil, visitor)
		if err != nil {
			return nil, err
		}
	} else {
		// NOTE: this is naive implementation, where for each point we re-walk the KD tree to intersect.
		// We could instead do a similar
		// optimization as the 1D case, but I think it'd mean building a query-time KD tree so we could
		// efficiently intersect against the
		// index, which is probably tricky!
		visitor := r.p.NewSinglePointVisitor(result)
		iterator, err := r.p.sortedPackedPoints.Iterator()
		if err != nil {
			return nil, err
		}

		for {
			point, err := iterator.Next(nil)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, err
			}
			visitor.SetPoint(point)
			err = values.Intersect(nil, visitor)
			if err != nil {
				return nil, err
			}
		}
	}
	return NewConstantScoreScorer(r, r.score, r.scoreMode, result.Build().Iterator())
}

func (r *pisQueryWeight) IsCacheable(ctx *index.LeafReaderContext) bool {
	return true
}

func (p *PointInSetQuery) Rewrite(reader index.Reader) (Query, error) {
	return p, nil
}

func (p *PointInSetQuery) Visit(visitor QueryVisitor) (err error) {
	if visitor.AcceptField(p.field) {
		return visitor.VisitLeaf(p)
	}
	return nil
}

var _ types.IntersectVisitor = &MergePointVisitor{}

type MergePointVisitor struct {
	result             *DocIdSetBuilder
	iterator           *index.TermIterator
	nextQueryPoint     []byte
	scratch            []byte
	sortedPackedPoints *index.PrefixCodedTerms
	adder              BulkAdder
}

func (p *PointInSetQuery) NewMergePointVisitor(sortedPackedPoints *index.PrefixCodedTerms,
	result *DocIdSetBuilder) (*MergePointVisitor, error) {
	iterator, err := sortedPackedPoints.Iterator()
	if err != nil {
		return nil, err
	}

	next, err := iterator.Next(nil)
	if err != nil {
		return nil, err
	}

	return &MergePointVisitor{
		result:             result,
		sortedPackedPoints: sortedPackedPoints,
		scratch:            make([]byte, p.bytesPerDim),
		iterator:           iterator,
		nextQueryPoint:     next,
	}, nil
}

func (m *MergePointVisitor) Visit(docID int) error {
	m.adder.Add(docID)
	return nil
}

func (m *MergePointVisitor) VisitLeaf(docID int, packedValue []byte) error {
	if m.matches(packedValue) {
		return m.Visit(docID)
	}
	return nil
}

func (m *MergePointVisitor) VisitIterator(iterator types.DocValuesIterator, packedValue []byte) error {
	if m.matches(packedValue) {
		for {
			docID, err := iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			err = m.Visit(docID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MergePointVisitor) matches(packedValue []byte) bool {
	m.scratch = packedValue
	for m.nextQueryPoint != nil {
		cmp := bytes.Compare(m.nextQueryPoint, m.scratch)
		switch {
		case cmp == 0:
			return true
		case cmp < 0:
			next, err := m.iterator.Next(nil)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return false
			}
			m.nextQueryPoint = next
		default:
			break
		}
	}
	return false
}

func (m *MergePointVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	var err error
	for m.nextQueryPoint != nil {
		cmpMin := bytes.Compare(m.nextQueryPoint, minPackedValue)
		if cmpMin < 0 {
			// query point is before the start of this cell
			m.nextQueryPoint, err = m.iterator.Next(nil)
			if err != nil {
				break
			}
			continue
		}

		cmpMax := bytes.Compare(m.nextQueryPoint, maxPackedValue)
		if cmpMax > 0 {
			// query point is after the end of this cell
			return types.CELL_OUTSIDE_QUERY
		}

		if cmpMin == 0 && cmpMax == 0 {
			// NOTE: we only hit this if we are on a cell whose min and max values are exactly equal to our point,
			// which can easily happen if many (> 1024) docs share this one value
			return types.CELL_INSIDE_QUERY
		} else {
			return types.CELL_CROSSES_QUERY
		}
	}

	// We exhausted all points in the query:
	return types.CELL_OUTSIDE_QUERY
}

func (m *MergePointVisitor) Grow(count int) {
	m.adder = m.result.Grow(count)
}

var _ types.IntersectVisitor = &SinglePointVisitor{}

type SinglePointVisitor struct {
	result     *DocIdSetBuilder
	pointBytes []byte
	adder      BulkAdder
	p          *PointInSetQuery
}

func (p *PointInSetQuery) NewSinglePointVisitor(result *DocIdSetBuilder) *SinglePointVisitor {
	return &SinglePointVisitor{
		result:     result,
		pointBytes: make([]byte, p.bytesPerDim*p.numDims),
		p:          p,
	}
}

func (s *SinglePointVisitor) SetPoint(point []byte) {
	copy(s.pointBytes, point)
}

func (s *SinglePointVisitor) Visit(docID int) error {
	s.adder.Add(docID)
	return nil
}

func (s *SinglePointVisitor) VisitLeaf(docID int, packedValue []byte) error {
	if bytes.Equal(packedValue, s.pointBytes) {
		// The point for this doc matches the point we are querying on
		return s.Visit(docID)
	}
	return nil
}

func (s *SinglePointVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	crosses := false

	bytesPerDim := s.p.bytesPerDim
	for dim := 0; dim < s.p.numDims; dim++ {
		offset := dim * bytesPerDim

		cmpMin := bytes.Compare(minPackedValue[offset:offset+bytesPerDim], s.pointBytes[offset:offset+bytesPerDim])
		if cmpMin > 0 {
			return types.CELL_OUTSIDE_QUERY
		}

		cmpMax := bytes.Compare(maxPackedValue[offset:offset+bytesPerDim], s.pointBytes[offset:offset+bytesPerDim])
		if cmpMax < 0 {
			return types.CELL_OUTSIDE_QUERY
		}

		if cmpMin != 0 || cmpMax != 0 {
			crosses = true
		}
	}

	if crosses {
		return types.CELL_CROSSES_QUERY
	}

	// NOTE: we only hit this if we are on a cell whose min and max values are exactly equal to our point,
	// which can easily happen if many docs share this one value
	return types.CELL_INSIDE_QUERY
}

func (s *SinglePointVisitor) Grow(count int) {
	s.adder = s.result.Grow(count)
}
