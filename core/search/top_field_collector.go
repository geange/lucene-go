package search

import (
	"context"
	"github.com/geange/lucene-go/core/index"
	. "github.com/geange/lucene-go/core/util/structure"
)

type MultiComparatorLeafCollector struct {
	comparator index.LeafFieldComparator
	reverseMul int
	scorer     Scorable
}

func (c *MultiComparatorLeafCollector) SetScorer(scorer Scorable) error {
	if err := c.comparator.SetScorer(scorer); err != nil {
		return err
	}
	c.scorer = scorer
	return nil
}

func NewMultiComparatorLeafCollector(comparators []index.LeafFieldComparator, reverseMul []int) *MultiComparatorLeafCollector {
	this := &MultiComparatorLeafCollector{}
	if len(comparators) == 1 {
		this.reverseMul = reverseMul[0]
		this.comparator = comparators[0]
	} else {
		this.reverseMul = 1
		this.comparator = NewMultiLeafFieldComparator(comparators, reverseMul)
	}
	return this
}

type TopFieldCollector struct {
	*TopDocsCollectorDefault[*Entry]

	numHits              int
	hitsThresholdChecker HitsThresholdChecker
	firstComparator      index.FieldComparator
	canSetMinScore       bool

	// shows if Search Sort if a part of the Index Sort
	searchSortPartOfIndexSort *Box[bool]

	// an accumulator that maintains the maximum of the segment's minimum competitive scores
	minScoreAcc *MaxScoreAccumulator

	// the current local minimum competitive score already propagated to the underlying scorer
	minCompetitiveScore float32

	numComparators int

	bottom *Entry

	queueFull bool

	docBase int

	needsScores bool

	scoreMode *ScoreMode
}

func (t *TopFieldCollector) ScoreMode() *ScoreMode {
	return t.scoreMode
}

func (t *TopFieldCollector) updateMinCompetitiveScore(scorer Scorable) error {
	if t.canSetMinScore && t.queueFull && t.hitsThresholdChecker.IsThresholdReached() {
		//assert bottom != null;
		minScore := t.firstComparator.Value(t.bottom.slot).(float32)
		if minScore > t.minCompetitiveScore {
			err := scorer.SetMinCompetitiveScore(minScore)
			if err != nil {
				return err
			}
			t.minCompetitiveScore = minScore
			t.totalHitsRelation = GREATER_THAN_OR_EQUAL_TO
			if t.minScoreAcc != nil {
				err := t.minScoreAcc.Accumulate(t.docBase, float32(minScore))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *TopFieldCollector) updateGlobalMinCompetitiveScore(scorer Scorable) error {
	if t.canSetMinScore && t.hitsThresholdChecker.IsThresholdReached() {
		// we can start checking the global maximum score even
		// if the local queue is not full because the threshold
		// is reached.
		maxMinScore := t.minScoreAcc.Get()
		if maxMinScore != nil && maxMinScore.score > t.minCompetitiveScore {
			scorer.SetMinCompetitiveScore(maxMinScore.score)
			t.minCompetitiveScore = maxMinScore.score
			t.totalHitsRelation = GREATER_THAN_OR_EQUAL_TO
		}
	}
	return nil
}

var _ TopDocsCollector = &SimpleFieldCollector{}

type SimpleFieldCollector struct {
	sort *index.Sort

	queue *PriorityQueue[*Entry]

	*TopFieldCollector
	*TopDocsCollectorDefault[*Entry]
}

func (s *SimpleFieldCollector) GetLeafCollector(ctx context.Context, leafCtx *index.LeafReaderContext) (LeafCollector, error) {
	// reset the minimum competitive score
	panic("")
}

type simpleLeafCollector struct {
	*MultiComparatorLeafCollector
}

var _ TopDocsCollector = &PagingFieldCollector{}

type PagingFieldCollector struct {
	sort          *index.Sort
	collectedHits int
	queue         *PriorityQueue[*Entry]
	after         *FieldDoc

	*TopFieldCollector
	*TopDocsCollectorDefault[*Entry]
}

func (p *PagingFieldCollector) GetLeafCollector(ctx context.Context, leafCtx *index.LeafReaderContext) (LeafCollector, error) {
	// reset the minimum competitive score
	p.minCompetitiveScore = 0
	p.docBase = leafCtx.DocBase
	afterDoc := p.after.doc - p.docBase
	// as all segments are sorted in the same way, enough to check only the 1st segment for indexSort
	if p.searchSortPartOfIndexSort == nil {
		indexSort := leafCtx.Reader().GetMetaData().GetSort()
		p.searchSortPartOfIndexSort = NewBox[bool](canEarlyTerminate(p.sort, indexSort))
		if p.searchSortPartOfIndexSort.Value() {
			p.firstComparator.DisableSkipping()
		}
	}

	return &pagingLeafCollector{}, nil
}

var _ LeafCollector = &pagingLeafCollector{}

type pagingLeafCollector struct {
	*MultiComparatorLeafCollector

	collectedAllCompetitiveHits bool
	tfc                         *TopFieldCollector
}

func (p *pagingLeafCollector) SetScorer(scorer Scorable) error {
	err := p.MultiComparatorLeafCollector.SetScorer(scorer)
	if err != nil {
		return err
	}
	if p.tfc.minScoreAcc == nil {
		err := p.tfc.updateMinCompetitiveScore(scorer)
		if err != nil {
			return err
		}
	} else {
		err := p.tfc.updateGlobalMinCompetitiveScore(scorer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *pagingLeafCollector) Collect(ctx context.Context, doc int) error {
	//TODO implement me
	panic("implement me")
}

func (p *pagingLeafCollector) CompetitiveIterator() (index.DocIdSetIterator, error) {
	//TODO implement me
	panic("implement me")
}

func canEarlyTerminate(searchSort, indexSort *index.Sort) bool {
	panic("")
}
