package search

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

type WeightScorer interface {
	Scorer(ctx index.LeafReaderContext) (index.Scorer, error)
}

type BaseWeight struct {
	scorer WeightScorer

	parentQuery index.Query
}

func NewBaseWeight(parentQuery index.Query, scorer WeightScorer) *BaseWeight {
	return &BaseWeight{
		scorer:      scorer,
		parentQuery: parentQuery,
	}
}

func (r *BaseWeight) GetQuery() index.Query {
	return r.parentQuery
}

//func (r *BaseWeight) IsCacheable(ctx index.LeafReaderContext) bool {
//	return false
//}

func (r *BaseWeight) Matches(ctx index.LeafReaderContext, doc int) (index.Matches, error) {
	supplier, err := r.ScorerSupplier(ctx)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, nil
	}

	scorer, err := supplier.Get(1)
	if err != nil {
		return nil, err
	}
	twoPhase := scorer.TwoPhaseIterator()
	if twoPhase == nil {
		advance, err := scorer.Iterator().Advance(nil, doc)
		if err != nil {
			return nil, err
		}
		if advance != doc {
			return nil, nil
		}
	} else {
		advance, err := twoPhase.Approximation().Advance(nil, doc)
		if err != nil {
			return nil, err
		}

		if ok, _ := twoPhase.Matches(); advance != doc || !ok {
			return nil, nil
		}
	}
	return nil, errors.New("MATCH_WITH_NO_TERMS")
}

func (r *BaseWeight) ScorerSupplier(ctx index.LeafReaderContext) (index.ScorerSupplier, error) {
	scorer, err := r.scorer.Scorer(ctx)
	if err != nil {
		return nil, err
	}
	if scorer == nil {
		return nil, nil
	}

	return &scorerSupplier{scorer: scorer}, nil
}

var _ index.ScorerSupplier = &scorerSupplier{}

type scorerSupplier struct {
	scorer index.Scorer
}

func (s *scorerSupplier) Get(leadCost int64) (index.Scorer, error) {
	return s.scorer, nil
}

func (s *scorerSupplier) Cost() int64 {
	return s.scorer.Iterator().Cost()
}

func (r *BaseWeight) BulkScorer(ctx index.LeafReaderContext) (index.BulkScorer, error) {
	scorer, err := r.scorer.Scorer(ctx)
	if err != nil {
		return nil, err
	}

	if scorer == nil {
		return nil, nil
	}

	return NewDefaultBulkScorer(scorer), nil
}

var _ index.BulkScorer = &DefaultBulkScorer{}

type DefaultBulkScorer struct {
	scorer   index.Scorer
	iterator types.DocIdSetIterator
	twoPhase index.TwoPhaseIterator
}

func NewDefaultBulkScorer(scorer index.Scorer) *DefaultBulkScorer {
	return &DefaultBulkScorer{
		scorer:   scorer,
		iterator: scorer.Iterator(),
		twoPhase: scorer.TwoPhaseIterator(),
	}
}

func (d *DefaultBulkScorer) Score(collector index.LeafCollector, acceptDocs util.Bits) error {
	NoMoreDocs := math.MaxInt32
	_, err := d.ScoreRange(collector, acceptDocs, 0, NoMoreDocs)
	return err
}

func (d *DefaultBulkScorer) ScoreRange(collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	err := collector.SetScorer(d.scorer)
	if err != nil {
		return 0, err
	}

	scorerIterator := func() types.DocIdSetIterator {
		if d.twoPhase == nil {
			return d.iterator
		}
		return d.twoPhase.Approximation()
	}()

	competitiveIterator, err := collector.CompetitiveIterator()
	if err != nil {
		return 0, err
	}

	var filteredIterator types.DocIdSetIterator
	if competitiveIterator == nil {
		filteredIterator = scorerIterator
	} else {
		// Wrap CompetitiveIterator and ScorerIterator start with (i.e., calling nextDoc()) the last
		// visited docID because ConjunctionDISI might have advanced to it in the previous
		// scoreRange, but we didn't process due to the range limit of scoreRange.
		if scorerIterator.DocID() != -1 {
			scorerIterator = NewStartDISIWrapper(scorerIterator)
		}

		if competitiveIterator.DocID() != -1 {
			competitiveIterator = NewStartDISIWrapper(competitiveIterator)
		}

		filteredIterator = IntersectIterators([]types.DocIdSetIterator{
			scorerIterator,
			competitiveIterator,
		})
	}

	if filteredIterator.DocID() == -1 && min == 0 && max == types.NO_MORE_DOCS {
		if err := scoreAll(collector, filteredIterator, d.twoPhase, acceptDocs); err != nil {
			return 0, err
		}
		return types.NO_MORE_DOCS, nil
	} else {
		doc := filteredIterator.DocID()
		if doc < min {
			doc, err = filteredIterator.Advance(nil, min)
			if err != nil {
				return 0, err
			}
		}
		return scoreRange(collector, filteredIterator, d.twoPhase, acceptDocs, doc, max)
	}
}

func scoreAll(collector index.LeafCollector, iterator types.DocIdSetIterator,
	twoPhase index.TwoPhaseIterator, acceptDocs util.Bits) error {

	doc, err := iterator.NextDoc()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	if twoPhase == nil {
		for {
			if acceptDocs == nil || acceptDocs.Test(uint(doc)) {
				err := collector.Collect(nil, doc)
				if err != nil {
					return err
				}
			}

			doc, err = iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		}
	} else {
		// The scorer has an approximation, so run the approximation first, then check acceptDocs, then confirm
		for {
			if ok, _ := twoPhase.Matches(); ok && (acceptDocs == nil || acceptDocs.Test(uint(doc))) {
				if err := collector.Collect(nil, doc); err != nil {
					return err
				}
			}

			doc, err = iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		}
	}

}

func scoreRange(collector index.LeafCollector, iterator types.DocIdSetIterator, twoPhase index.TwoPhaseIterator,
	acceptDocs util.Bits, currentDoc, end int) (int, error) {

	var err error

	if twoPhase == nil {
		for currentDoc < end {
			if acceptDocs == nil || acceptDocs.Test(uint(currentDoc)) {
				err := collector.Collect(nil, currentDoc)
				if err != nil {
					return 0, err
				}
			}
			currentDoc, err = iterator.NextDoc()
			if err != nil {
				return 0, err
			}
		}
		return currentDoc, nil
	} else {
		for currentDoc < end {
			if ok, _ := twoPhase.Matches(); ok && (acceptDocs == nil || acceptDocs.Test(uint(currentDoc))) {
				err := collector.Collect(nil, currentDoc)
				if err != nil {
					return 0, err
				}
			}

			currentDoc, err = iterator.NextDoc()
			if err != nil {
				return 0, err
			}
		}
		return currentDoc, nil
	}
}

func (d *DefaultBulkScorer) Cost() int64 {
	return d.iterator.Cost()
}

var _ types.DocIdSetIterator = &StartDISIWrapper{}

type StartDISIWrapper struct {
	in         types.DocIdSetIterator
	startDocID int
	docID      int
}

func NewStartDISIWrapper(in types.DocIdSetIterator) *StartDISIWrapper {
	return &StartDISIWrapper{
		in:         in,
		startDocID: in.DocID(),
	}
}

func (s *StartDISIWrapper) DocID() int {
	return s.docID
}

func (s *StartDISIWrapper) NextDoc() (int, error) {
	return s.Advance(nil, s.docID+1)
}

func (s *StartDISIWrapper) Advance(ctx context.Context, target int) (int, error) {
	if target <= s.startDocID {
		s.docID = s.startDocID
		return s.docID, nil
	}
	var err error
	s.docID, err = s.in.Advance(nil, target)
	return s.docID, err
}

func (s *StartDISIWrapper) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvance(s, target)
}

func (s *StartDISIWrapper) Cost() int64 {
	return s.in.Cost()
}
