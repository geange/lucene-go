package search

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"math"

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

type BaseTopScoreDocCollector struct {
	*TopDocsCollectorDefault[ScoreDoc]

	docBase              int
	pqTop                ScoreDoc
	hitsThresholdChecker HitsThresholdChecker
	minScoreAcc          *MaxScoreAccumulator
	minCompetitiveScore  float64
}

func (t *BaseTopScoreDocCollector) updateGlobalMinCompetitiveScore(scorer Scorable) error {
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

func (t *BaseTopScoreDocCollector) updateMinCompetitiveScore(scorer Scorable) error {
	if t.hitsThresholdChecker.IsThresholdReached() && t.pqTop != nil && !math.IsInf(t.pqTop.GetScore(), -1) { // -Infinity is the score of sentinels

		// since we tie-break on doc id and collect in doc id order, we can require
		// the next float
		localMinScore := t.pqTop.GetScore()
		if localMinScore > t.minCompetitiveScore {
			if err := scorer.SetMinCompetitiveScore(localMinScore); err != nil {
				return err
			}
			t.totalHitsRelation = GREATER_THAN_OR_EQUAL_TO
			t.minCompetitiveScore = localMinScore
			if t.minScoreAcc != nil {
				// we don't use the next float but we register the document
				// id so that other leaves can require it if they are after
				// the current maximum
				if err := t.minScoreAcc.Accumulate(t.docBase, float32(t.pqTop.GetScore())); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func newTopScoreDocCollector(numHits int, hitsThresholdChecker HitsThresholdChecker,
	minScoreAcc *MaxScoreAccumulator) *BaseTopScoreDocCollector {

	ts := &BaseTopScoreDocCollector{
		hitsThresholdChecker: hitsThresholdChecker,
		minScoreAcc:          minScoreAcc,
	}

	queue := structure.NewPriorityQueueV1(numHits,
		func() ScoreDoc {
			return newScoreDoc(math.MaxInt32, math.Inf(-1))
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

type ScorerLeafCollector struct {
	p      *BaseTopScoreDocCollector
	scorer Scorable
}

func (s *ScorerLeafCollector) SetScorer(scorer Scorable) error {
	s.scorer = scorer
	return nil
}

var _ TopScoreDocCollector = &SimpleTopScoreDocCollector{}

type SimpleTopScoreDocCollector struct {
	*BaseTopScoreDocCollector
}

func NewSimpleTopScoreDocCollector(numHits int, hitsThresholdChecker HitsThresholdChecker,
	minScoreAcc *MaxScoreAccumulator) (TopScoreDocCollector, error) {
	return &SimpleTopScoreDocCollector{
		newTopScoreDocCollector(numHits, hitsThresholdChecker, minScoreAcc),
	}, nil
}

var _ LeafCollector = &simpleTopScoreDocCollectorLeafCollector{}

type simpleTopScoreDocCollectorLeafCollector struct {
	*ScorerLeafCollector
}

func (s *simpleTopScoreDocCollectorLeafCollector) Collect(ctx context.Context, doc int) error {
	score, err := s.scorer.Score()
	if err != nil {
		return err
	}

	// This collector relies on the fact that scorers produce positive values:
	// assert score >= 0; // NOTE: false for NaN
	s.p.totalHits++
	s.p.hitsThresholdChecker.IncrementHitCount()

	if s.p.minScoreAcc != nil && (int64(s.p.totalHits)&s.p.minScoreAcc.modInterval) == 0 {
		if err := s.p.updateGlobalMinCompetitiveScore(s.scorer); err != nil {
			return err
		}
	}

	if score <= s.p.pqTop.GetScore() {
		if s.p.totalHitsRelation == EQUAL_TO {
			// we just reached totalHitsThreshold, we can start setting the min
			// competitive score now
			if err := s.p.updateMinCompetitiveScore(s.scorer); err != nil {
				return err
			}
		}
		// Since docs are returned in-order (i.e., increasing doc Id), a document
		// with equal score to pqTop.score cannot compete since HitQueue favors
		// documents with lower doc Ids. Therefore reject those docs too.
		return nil
	}

	s.p.pqTop.SetDoc(doc + s.p.docBase)
	s.p.pqTop.SetScore(score)
	s.p.pqTop = s.p.pq.UpdateTop()
	return s.p.updateMinCompetitiveScore(s.scorer)
}

func (s *simpleTopScoreDocCollectorLeafCollector) SetScorer(scorer Scorable) error {
	if err := s.ScorerLeafCollector.SetScorer(scorer); err != nil {
		return err
	}

	if s.p.minScoreAcc == nil {
		return s.p.updateMinCompetitiveScore(s.scorer)
	}
	return s.p.updateGlobalMinCompetitiveScore(s.scorer)
}

func (s *simpleTopScoreDocCollectorLeafCollector) CompetitiveIterator() (types.DocIdSetIterator, error) {
	return nil, nil
}

func (s *SimpleTopScoreDocCollector) GetLeafCollector(ctx context.Context, readerContext index.LeafReaderContext) (LeafCollector, error) {
	// reset the minimum competitive score
	s.minCompetitiveScore = 0
	s.docBase = readerContext.DocBase()

	return &simpleTopScoreDocCollectorLeafCollector{&ScorerLeafCollector{
		p: s.BaseTopScoreDocCollector,
	}}, nil

}

func (s *SimpleTopScoreDocCollector) ScoreMode() ScoreMode {
	return s.hitsThresholdChecker.ScoreMode()
}

var _ TopScoreDocCollector = &PagingTopScoreDocCollector{}

type PagingTopScoreDocCollector struct {
	*BaseTopScoreDocCollector

	after         ScoreDoc
	collectedHits int
}

var _ LeafCollector = &pagingTopScoreDocCollectorLeafCollector{}

type pagingTopScoreDocCollectorLeafCollector struct {
	*PagingTopScoreDocCollector

	scorer   Scorable
	afterDoc int
}

func (p *pagingTopScoreDocCollectorLeafCollector) SetScorer(scorer Scorable) error {
	p.scorer = scorer
	if p.minScoreAcc == nil {
		return p.updateMinCompetitiveScore(scorer)
	}
	return p.updateGlobalMinCompetitiveScore(scorer)
}

func (p *pagingTopScoreDocCollectorLeafCollector) Collect(ctx context.Context, doc int) error {
	score, err := p.scorer.Score()
	if err != nil {
		return err
	}

	// This collector relies on the fact that scorers produce positive values:
	// assert score >= 0; // NOTE: false for NaN
	p.totalHits++
	p.hitsThresholdChecker.IncrementHitCount()

	if p.minScoreAcc != nil && (int64(p.totalHits)&p.minScoreAcc.modInterval) == 0 {
		err := p.updateGlobalMinCompetitiveScore(p.scorer)
		if err != nil {
			return err
		}
	}

	if score > p.after.GetScore() || (score == p.after.GetScore() && doc <= p.afterDoc) {
		// hit was collected on a previous page
		if p.totalHitsRelation == EQUAL_TO {
			// we just reached totalHitsThreshold, we can start setting the min
			// competitive score now
			if err := p.updateMinCompetitiveScore(p.scorer); err != nil {
				return err
			}
		}
		return nil
	}

	if score <= p.pqTop.GetScore() {
		if p.totalHitsRelation == EQUAL_TO {
			// we just reached totalHitsThreshold, we can start setting the min
			// competitive score now
			if err := p.updateMinCompetitiveScore(p.scorer); err != nil {
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
	p.pqTop.SetScore(score)
	p.pqTop = p.pq.UpdateTop()
	return p.updateMinCompetitiveScore(p.scorer)
}

func (p *pagingTopScoreDocCollectorLeafCollector) CompetitiveIterator() (types.DocIdSetIterator, error) {
	return nil, nil
}

func (p *PagingTopScoreDocCollector) GetLeafCollector(ctx context.Context, readerContext index.LeafReaderContext) (LeafCollector, error) {
	// reset the minimum competitive score
	p.minCompetitiveScore = 0
	p.docBase = readerContext.DocBase()
	afterDoc := p.after.GetDoc() - readerContext.DocBase()

	return &pagingTopScoreDocCollectorLeafCollector{
		PagingTopScoreDocCollector: p,
		scorer:                     nil,
		afterDoc:                   afterDoc,
	}, nil
}

func (p *PagingTopScoreDocCollector) ScoreMode() ScoreMode {
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
