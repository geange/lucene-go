package search

import (
	"context"
	"errors"
	"math"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util/structure"
)

// TopScoreDocCollector
// A Collector implementation that collects the top-scoring hits,
// returning them as a TopDocs. This is used by IndexSearcher to implement TopDocs-based search.
// Hits are sorted by score descending and then (when the scores are tied) docID ascending.
// When you create an instance of this collector you should know in advance whether documents
// are going to be collected in doc Id order or not.
//
// NOTE: The values Float.NaN and Float.NEGATIVE_INFINITY are not valid scores. This collector
// will not properly collect hits with such scores.
type TopScoreDocCollector interface {
	TopDocsCollector
}

type TopScoreDocCollectorDefault struct {
	*TopDocsCollectorDefault[ScoreDoc]
	docBase              int
	pqTop                ScoreDoc
	hitsThresholdChecker HitsThresholdChecker
	minScoreAcc          *MaxScoreAccumulator
	minCompetitiveScore  float32
}

func (t *TopScoreDocCollectorDefault) updateGlobalMinCompetitiveScore(scorer Scorable) error {
	//assert minScoreAcc != null;
	maxMinScore := t.minScoreAcc.Get()
	if maxMinScore != nil {
		// since we tie-break on doc id and collect in doc id order we can require
		// the next float if the global minimum score is set on a document id that is
		// smaller than the ids in the current leaf
		score := maxMinScore.score
		if t.docBase >= maxMinScore.docBase {
			score = maxMinScore.score
		}
		if score > t.minCompetitiveScore {
			//assert hitsThresholdChecker.isThresholdReached();
			if err := scorer.SetMinCompetitiveScore(score); err != nil {
				return err
			}
			t.minCompetitiveScore = score
			t.totalHitsRelation = GREATER_THAN_OR_EQUAL_TO
		}
	}
	return nil
}

func (t *TopScoreDocCollectorDefault) updateMinCompetitiveScore(scorer Scorable) error {
	if t.hitsThresholdChecker.IsThresholdReached() &&
		t.pqTop != nil &&
		!math.IsInf(t.pqTop.GetScore(), -1) { // -Infinity is the score of sentinels

		// since we tie-break on doc id and collect in doc id order, we can require
		// the next float
		localMinScore := t.pqTop.GetScore()
		if localMinScore > float64(t.minCompetitiveScore) {
			err := scorer.SetMinCompetitiveScore(float32(localMinScore))
			if err != nil {
				return err
			}
			t.totalHitsRelation = GREATER_THAN_OR_EQUAL_TO
			t.minCompetitiveScore = float32(localMinScore)
			if t.minScoreAcc != nil {
				// we don't use the next float but we register the document
				// id so that other leaves can require it if they are after
				// the current maximum
				err := t.minScoreAcc.Accumulate(t.docBase, float32(t.pqTop.GetScore()))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func newTopScoreDocCollector(numHits int,
	hitsThresholdChecker HitsThresholdChecker,
	minScoreAcc *MaxScoreAccumulator) *TopScoreDocCollectorDefault {

	ts := &TopScoreDocCollectorDefault{
		hitsThresholdChecker: hitsThresholdChecker,
		minScoreAcc:          minScoreAcc,
	}

	queue := structure.NewPriorityQueueV1(numHits,
		func() ScoreDoc {
			return NewScoreDoc(math.MaxInt32, math.Inf(-1))
		},
		func(hitA, hitB ScoreDoc) bool {
			if hitA.GetScore() == hitB.GetScore() {
				return hitA.GetDoc() > hitB.GetDoc()
			}
			return hitA.GetScore() < hitB.GetScore()
		})
	queue.SetSize(numHits)

	ts.TopDocsCollectorDefault = newTopDocsCollectorDefault(queue)
	// HitQueue implements getSentinelObject to return a ScoreDoc, so we know
	// that at this point top() is already initialized.
	ts.pqTop = ts.pq.Top()

	return ts
}

func TopScoreDocCollectorCreate(numHits int, after ScoreDoc,
	hitsThresholdChecker HitsThresholdChecker, minScoreAcc *MaxScoreAccumulator) (TopScoreDocCollector, error) {

	if numHits <= 0 {
		return nil, errors.New("numHits must be > 0; please use TotalHitCountCollector if you just need the total hit count")
	}

	if hitsThresholdChecker == nil {
		return nil, errors.New("hitsThresholdChecker must be non null")
	}

	if after == nil {
		return NewSimpleTopScoreDocCollector(numHits, hitsThresholdChecker, minScoreAcc)
	}
	return NewPagingTopScoreDocCollector(numHits, after, hitsThresholdChecker, minScoreAcc)
}

var _ LeafCollector = &ScorerLeafCollector{}

type ScorerLeafCollector struct {
	scorer Scorable

	FnSetScorer func(scorer Scorable) error
	FnCollect   func(ctx context.Context, doc int) error
}

func (s *ScorerLeafCollector) Collect(ctx context.Context, doc int) error {
	return s.FnCollect(ctx, doc)
}

func (s *ScorerLeafCollector) CompetitiveIterator() (index.DocIdSetIterator, error) {
	return nil, nil
}

func (s *ScorerLeafCollector) SetScorer(scorer Scorable) error {
	s.scorer = scorer

	if s.FnSetScorer != nil {
		return s.FnSetScorer(scorer)
	}
	return nil
}

var _ TopScoreDocCollector = &SimpleTopScoreDocCollector{}

type SimpleTopScoreDocCollector struct {
	*TopScoreDocCollectorDefault
}

func NewSimpleTopScoreDocCollector(numHits int, hitsThresholdChecker HitsThresholdChecker,
	minScoreAcc *MaxScoreAccumulator) (TopScoreDocCollector, error) {
	return &SimpleTopScoreDocCollector{
		newTopScoreDocCollector(numHits, hitsThresholdChecker, minScoreAcc),
	}, nil
}

func (s *SimpleTopScoreDocCollector) GetLeafCollector(_ context.Context,
	readerContext *index.LeafReaderContext) (LeafCollector, error) {
	// reset the minimum competitive score
	s.minCompetitiveScore = 0
	s.docBase = readerContext.DocBase

	c := &ScorerLeafCollector{
		FnSetScorer: func(scorer Scorable) error {
			if s.minScoreAcc == nil {
				return s.updateMinCompetitiveScore(scorer)
			}
			return s.updateGlobalMinCompetitiveScore(scorer)
		},
	}

	c.FnCollect = func(ctx context.Context, doc int) error {
		score, err := c.scorer.Score()
		if err != nil {
			return err
		}

		// This collector relies on the fact that scorers produce positive values:
		// assert score >= 0; // NOTE: false for NaN
		s.totalHits++
		s.hitsThresholdChecker.IncrementHitCount()

		if s.minScoreAcc != nil && (int64(s.totalHits)&s.minScoreAcc.modInterval) == 0 {
			if err := s.updateGlobalMinCompetitiveScore(c.scorer); err != nil {
				return err
			}
		}

		if float64(score) <= s.pqTop.GetScore() {
			if s.totalHitsRelation == EQUAL_TO {
				// we just reached totalHitsThreshold, we can start setting the min
				// competitive score now
				if err := s.updateMinCompetitiveScore(c.scorer); err != nil {
					return err
				}
			}
			// Since docs are returned in-order (i.e., increasing doc Id), a document
			// with equal score to pqTop.score cannot compete since HitQueue favors
			// documents with lower doc Ids. Therefore reject those docs too.
			return nil
		}

		s.pqTop.SetDoc(doc + s.docBase)
		s.pqTop.SetScore(float64(score))
		s.pqTop = s.pq.UpdateTop()
		return s.updateMinCompetitiveScore(c.scorer)
	}
	return c, nil

}

func (s *SimpleTopScoreDocCollector) ScoreMode() *ScoreMode {
	return s.hitsThresholdChecker.ScoreMode()
}

var _ TopScoreDocCollector = &PagingTopScoreDocCollector{}

type PagingTopScoreDocCollector struct {
	*TopScoreDocCollectorDefault
	after         ScoreDoc
	collectedHits int
}

func (p *PagingTopScoreDocCollector) GetLeafCollector(ctx context.Context, leafCtx *index.LeafReaderContext) (LeafCollector, error) {
	// reset the minimum competitive score
	p.minCompetitiveScore = 0
	p.docBase = leafCtx.DocBase
	afterDoc := p.after.GetDoc() - leafCtx.DocBase

	c := &ScorerLeafCollector{}
	c.FnSetScorer = func(scorer Scorable) error {
		if p.minScoreAcc == nil {
			return p.updateMinCompetitiveScore(scorer)
		}
		return p.updateGlobalMinCompetitiveScore(scorer)
	}
	c.FnCollect = func(ctx context.Context, doc int) error {
		score, err := c.scorer.Score()
		if err != nil {
			return err
		}

		// This collector relies on the fact that scorers produce positive values:
		// assert score >= 0; // NOTE: false for NaN
		p.totalHits++
		p.hitsThresholdChecker.IncrementHitCount()

		if p.minScoreAcc != nil && (int64(p.totalHits)&p.minScoreAcc.modInterval) == 0 {
			err := p.updateGlobalMinCompetitiveScore(c.scorer)
			if err != nil {
				return err
			}
		}

		if float64(score) > p.after.GetScore() || (float64(score) == p.after.GetScore() && doc <= afterDoc) {
			// hit was collected on a previous page
			if p.totalHitsRelation == EQUAL_TO {
				// we just reached totalHitsThreshold, we can start setting the min
				// competitive score now
				err := p.updateMinCompetitiveScore(c.scorer)
				if err != nil {
					return err
				}
			}
			return nil
		}

		if float64(score) <= p.pqTop.GetScore() {
			if p.totalHitsRelation == EQUAL_TO {
				// we just reached totalHitsThreshold, we can start setting the min
				// competitive score now
				err := p.updateMinCompetitiveScore(c.scorer)
				if err != nil {
					return err
				}
			}

			// Since docs are returned in-order (i.e., increasing doc Id), a document
			// with equal score to pqTop.score cannot compete since HitQueue favors
			// documents with lower doc Ids. Therefore reject those docs too.
			return nil
		}

		p.collectedHits++
		p.pqTop.SetDoc(doc + p.docBase)
		p.pqTop.SetScore(float64(score))
		p.pqTop = p.pq.UpdateTop()
		return p.updateMinCompetitiveScore(c.scorer)
	}
	return c, nil
}

func (p *PagingTopScoreDocCollector) ScoreMode() *ScoreMode {
	//TODO implement me
	panic("implement me")
}

func (p *PagingTopScoreDocCollector) TopDocsSize() int {
	size := p.pq.Size()
	if p.collectedHits < p.pq.Size() {
		size = p.collectedHits
	}
	return size
}

func (p *PagingTopScoreDocCollector) NewTopDocs(results []ScoreDoc, howMany int) (TopDocs, error) {
	if len(results) != 0 {
		return NewTopDocs(NewTotalHits(int64(p.totalHits), p.totalHitsRelation), results), nil
	}
	return NewTopDocs(NewTotalHits(int64(p.totalHits), p.totalHitsRelation), make([]ScoreDoc, 0)), nil
}

func NewPagingTopScoreDocCollector(hits int, after ScoreDoc, checker HitsThresholdChecker, acc *MaxScoreAccumulator) (TopScoreDocCollector, error) {
	panic("")
}
