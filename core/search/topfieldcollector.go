package search

import (
	"context"
	"errors"
	"fmt"
	cindex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
	. "github.com/geange/lucene-go/core/util/structure"
)

func TopTopFieldCollectorCreate(sort index.Sort, numHits int, after FieldDoc,
	hitsThresholdChecker HitsThresholdChecker, minScoreAcc *MaxScoreAccumulator) (TopDocsCollector, error) {

	if len(sort.GetSort()) == 0 {
		return nil, errors.New("sort must contain at least one field")
	}

	if numHits <= 0 {
		return nil, errors.New("numHits must be > 0; please use TotalHitCountCollector if you just need the total hit count")
	}

	if hitsThresholdChecker == nil {
		return nil, errors.New("hitsThresholdChecker should not be null")
	}

	// here we assume that if hitsThreshold was set, we let a comparator to skip non-competitive docs
	queue := CreateFieldValueHitQueue(sort.GetSort(), numHits)
	if after == nil {
		// inform a comparator that sort is based on this single field
		// to enable some optimizations for skipping over non-competitive documents
		// We can't set single sort when the `after` parameter is non-null as it's
		// an implicit sort over the document id.
		if len(queue.GetComparatorsList()) == 1 {
			queue.GetComparatorsList()[0].SetSingleSort()
		}
		return NewSimpleFieldCollector(sort, queue, numHits, hitsThresholdChecker, minScoreAcc)
	} else {
		if len(after.GetFields()) == 0 {
			return nil, errors.New("after.fields wasn't set; you must pass fillFields=true for the previous search")
		}

		if len(after.GetFields()) != len(sort.GetSort()) {
			return nil, fmt.Errorf("after.fields has %d values but sort has %d", len(after.GetFields()), len(sort.GetSort()))
		}
		return NewPagingFieldCollector(sort, queue, after, numHits, hitsThresholdChecker, minScoreAcc)
	}
}

type MultiComparatorLeafCollector struct {
	comparator index.LeafFieldComparator
	reverseMul int
	scorer     search.Scorable
}

func (c *MultiComparatorLeafCollector) SetScorer(scorer search.Scorable) error {
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
	minCompetitiveScore float64

	numComparators int

	bottom *Entry

	queueFull bool

	docBase int

	needsScores bool

	scoreMode search.ScoreMode
}

func (t *TopFieldCollector) ScoreMode() search.ScoreMode {
	return t.scoreMode
}

func (t *TopFieldCollector) updateMinCompetitiveScore(scorer search.Scorable) error {
	if t.canSetMinScore && t.queueFull && t.hitsThresholdChecker.IsThresholdReached() {
		//assert bottom != null;
		minScore := t.firstComparator.Value(t.bottom.slot).(float64)
		if minScore > t.minCompetitiveScore {
			err := scorer.SetMinCompetitiveScore(minScore)
			if err != nil {
				return err
			}
			t.minCompetitiveScore = minScore
			t.totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
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

func (t *TopFieldCollector) updateGlobalMinCompetitiveScore(scorer search.Scorable) error {
	if t.canSetMinScore && t.hitsThresholdChecker.IsThresholdReached() {
		// we can start checking the global maximum score even
		// if the local queue is not full because the threshold
		// is reached.
		maxMinScore := t.minScoreAcc.Get()
		if maxMinScore != nil && maxMinScore.score > t.minCompetitiveScore {
			scorer.SetMinCompetitiveScore(maxMinScore.score)
			t.minCompetitiveScore = maxMinScore.score
			t.totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
		}
	}
	return nil
}

var _ TopDocsCollector = &SimpleFieldCollector{}

type SimpleFieldCollector struct {
	sort index.Sort

	queue FieldValueHitQueue[*Entry]

	*TopFieldCollector
	*TopDocsCollectorDefault[*Entry]
}

func NewSimpleFieldCollector(sort index.Sort, queue FieldValueHitQueue[*Entry], numHits int,
	hitsThresholdChecker HitsThresholdChecker, minScoreAcc *MaxScoreAccumulator) (*SimpleFieldCollector, error) {
	panic("")
}

func (s *SimpleFieldCollector) GetLeafCollector(ctx context.Context, readerContext index.LeafReaderContext) (search.LeafCollector, error) {
	// reset the minimum competitive score
	s.minCompetitiveScore = 0
	s.docBase = readerContext.DocBase()

	// as all segments are sorted in the same way, enough to check only the 1st segment for indexSort
	if s.searchSortPartOfIndexSort == nil {
		indexSort := readerContext.Reader().GetMetaData().GetSort()
		can := canEarlyTerminate(s.sort, indexSort)
		s.searchSortPartOfIndexSort = NewBox(can)
		if s.searchSortPartOfIndexSort.Value() {
			s.firstComparator.DisableSkipping()
		}
	}

	comparators, err := s.queue.GetComparators(readerContext)
	if err != nil {
		return nil, err
	}
	reverseMul := s.queue.GetReverseMul()

	return &simpleLeafCollector{
		SimpleFieldCollector:         s,
		MultiComparatorLeafCollector: NewMultiComparatorLeafCollector(comparators, reverseMul),
		collectedAllCompetitiveHits:  false,
	}, nil

}

var _ search.LeafCollector = &simpleLeafCollector{}

type simpleLeafCollector struct {
	*MultiComparatorLeafCollector
	*SimpleFieldCollector

	collectedAllCompetitiveHits bool
}

func (s *simpleLeafCollector) Collect(ctx context.Context, doc int) error {
	s.totalHits++
	s.hitsThresholdChecker.IncrementHitCount()

	if s.minScoreAcc != nil && (int64(s.totalHits)&s.minScoreAcc.modInterval) == 0 {
		if err := s.updateGlobalMinCompetitiveScore(s.scorer); err != nil {
			return err
		}
	}

	if s.scoreMode.IsExhaustive() == false && s.totalHitsRelation == search.EQUAL_TO &&
		s.hitsThresholdChecker.IsThresholdReached() {
		// for the first time hitsThreshold is reached, notify comparator about this
		if err := s.comparator.SetHitsThresholdReached(); err != nil {
			return err
		}
		s.totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
	}

	if s.queueFull {
		bottom, err := s.comparator.CompareBottom(doc)
		if err != nil {
			return err
		}
		if s.collectedAllCompetitiveHits || s.reverseMul*bottom <= 0 {
			// since docs are visited in doc Id order, if compare is 0, it means
			// this document is largest than anything else in the queue, and
			// therefore not competitive.
			if s.searchSortPartOfIndexSort.Value() {
				if s.hitsThresholdChecker.IsThresholdReached() {
					s.totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
					return errors.New("CollectionTerminatedException")
				} else {
					s.collectedAllCompetitiveHits = true
				}
			} else if s.totalHitsRelation == search.EQUAL_TO {
				// we can start setting the min competitive score if the
				// threshold is reached for the first time here.
				if err := s.updateMinCompetitiveScore(s.scorer); err != nil {
					return err
				}
			}
			return nil
		}

		// This hit is competitive - replace bottom element in queue & adjustTop
		if err := s.comparator.Copy(s.bottom.slot, doc); err != nil {
			return err
		}
		s.updateBottom(doc)
		if err := s.comparator.SetBottom(s.bottom.slot); err != nil {
			return err
		}
		return s.updateMinCompetitiveScore(s.scorer)
	} else {
		// Startup transient: queue hasn't gathered numHits yet
		slot := s.totalHits - 1

		// Copy hit into queue
		if err := s.comparator.Copy(slot, doc); err != nil {
			return err
		}
		s.add(slot, doc)
		if s.queueFull {
			if err := s.comparator.SetBottom(s.bottom.slot); err != nil {
				return err
			}
			if err := s.updateMinCompetitiveScore(s.scorer); err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *simpleLeafCollector) CompetitiveIterator() (types.DocIdSetIterator, error) {
	return s.comparator.CompetitiveIterator()
}

func (s *simpleLeafCollector) add(slot int, doc int) {
	s.bottom = s.pq.Add(NewEntry(slot, s.docBase+doc))
	s.queueFull = s.totalHits == s.numHits
}

var _ TopDocsCollector = &PagingFieldCollector{}

type PagingFieldCollector struct {
	*TopFieldCollector
	*TopDocsCollectorDefault[*Entry]

	sort          index.Sort
	collectedHits int
	queue         *FieldValueHitQueueDefault[*Entry]
	after         FieldDoc
}

func NewPagingFieldCollector(sort index.Sort, queue FieldValueHitQueue[*Entry], after FieldDoc, numHits int,
	hitsThresholdChecker HitsThresholdChecker, minScoreAcc *MaxScoreAccumulator) (*PagingFieldCollector, error) {

	panic("")
}

func (p *PagingFieldCollector) GetLeafCollector(ctx context.Context, readerContext index.LeafReaderContext) (search.LeafCollector, error) {
	// reset the minimum competitive score
	p.minCompetitiveScore = 0
	p.docBase = readerContext.DocBase()
	afterDoc := p.after.GetDoc() - p.docBase
	// as all segments are sorted in the same way, enough to check only the 1st segment for indexSort
	if p.searchSortPartOfIndexSort == nil {
		indexSort := readerContext.Reader().GetMetaData().GetSort()
		p.searchSortPartOfIndexSort = NewBox[bool](canEarlyTerminate(p.sort, indexSort))
		if p.searchSortPartOfIndexSort.Value() {
			p.firstComparator.DisableSkipping()
		}
	}

	comparators, err := p.queue.GetComparators(readerContext)
	if err != nil {
		return nil, err
	}
	reverseMul := p.queue.GetReverseMul()

	return &pagingLeafCollector{
		PagingFieldCollector:         p,
		MultiComparatorLeafCollector: NewMultiComparatorLeafCollector(comparators, reverseMul),
		collectedAllCompetitiveHits:  false,
		afterDoc:                     afterDoc,
	}, nil
}

var _ search.LeafCollector = &pagingLeafCollector{}

type pagingLeafCollector struct {
	*MultiComparatorLeafCollector
	*PagingFieldCollector

	collectedAllCompetitiveHits bool
	afterDoc                    int
}

func (p *pagingLeafCollector) SetScorer(scorer search.Scorable) error {
	if err := p.MultiComparatorLeafCollector.SetScorer(scorer); err != nil {
		return err
	}

	if p.minScoreAcc == nil {
		if err := p.updateMinCompetitiveScore(scorer); err != nil {
			return err
		}
	} else {
		if err := p.updateGlobalMinCompetitiveScore(scorer); err != nil {
			return err
		}
	}
	return nil
}

func (p *pagingLeafCollector) Collect(ctx context.Context, doc int) error {
	//System.out.println("  collect doc=" + doc);

	p.totalHits++
	p.hitsThresholdChecker.IncrementHitCount()

	if p.minScoreAcc != nil && (int64(p.totalHits)&p.minScoreAcc.modInterval) == 0 {
		if err := p.updateGlobalMinCompetitiveScore(p.scorer); err != nil {
			return err
		}
	}

	if p.scoreMode.IsExhaustive() == false && p.totalHitsRelation == search.EQUAL_TO &&
		p.hitsThresholdChecker.IsThresholdReached() {
		// for the first time hitsThreshold is reached, notify comparator about this
		if err := p.comparator.SetHitsThresholdReached(); err != nil {
			return err
		}
		p.totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
	}

	if p.queueFull {
		// Fastmatch: return if this hit is no better than
		// the worst hit currently in the queue:
		bottom, err := p.comparator.CompareBottom(doc)
		if err != nil {
			return err
		}
		if p.collectedAllCompetitiveHits || p.reverseMul*bottom <= 0 {
			// since docs are visited in doc Id order, if compare is 0, it means
			// this document is largest than anything else in the queue, and
			// therefore not competitive.
			if p.searchSortPartOfIndexSort.Value() {
				if p.hitsThresholdChecker.IsThresholdReached() {
					p.totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
					return errors.New("CollectionTerminatedException")
				} else {
					p.collectedAllCompetitiveHits = true
				}
			} else if p.totalHitsRelation == search.EQUAL_TO {
				// we can start setting the min competitive score if the
				// threshold is reached for the first time here.
				if err := p.updateMinCompetitiveScore(p.scorer); err != nil {
					return err
				}
			}
			return nil
		}
	}

	compareTop, err := p.comparator.CompareTop(doc)
	if err != nil {
		return err
	}
	topCmp := p.reverseMul * compareTop
	if topCmp > 0 || (topCmp == 0 && doc <= p.afterDoc) {
		// Already collected on a previous page
		if p.totalHitsRelation == search.EQUAL_TO {
			// check if totalHitsThreshold is reached and we can update competitive score
			// necessary to account for possible update to global min competitive score
			if err := p.updateMinCompetitiveScore(p.scorer); err != nil {
				return err
			}
		}
		return nil
	}

	if p.queueFull {
		// This hit is competitive - replace bottom element in queue & adjustTop
		if err := p.comparator.Copy(p.bottom.slot, doc); err != nil {
			return err
		}

		p.updateBottom(doc)

		if err := p.comparator.SetBottom(p.bottom.slot); err != nil {
			return err
		}
		return p.updateMinCompetitiveScore(p.scorer)
	} else {
		p.collectedHits++

		// Startup transient: queue hasn't gathered numHits yet
		slot := p.collectedHits - 1
		//System.out.println("    slot=" + slot);
		// Copy hit into queue
		if err := p.comparator.Copy(slot, doc); err != nil {
			return err
		}

		p.bottom = p.pq.Add(NewEntry(slot, p.docBase+doc))
		p.queueFull = p.collectedHits == p.numHits
		if p.queueFull {
			if err := p.comparator.SetBottom(p.bottom.slot); err != nil {
				return err
			}
			return p.updateMinCompetitiveScore(p.scorer)
		}
		return nil
	}
}

func (p *pagingLeafCollector) CompetitiveIterator() (types.DocIdSetIterator, error) {
	return p.comparator.CompetitiveIterator()
}

func canEarlyTerminate(searchSort, indexSort index.Sort) bool {
	return canEarlyTerminateOnDocId(indexSort) ||
		canEarlyTerminateOnPrefix(searchSort, indexSort)
}

func canEarlyTerminateOnDocId(searchSort index.Sort) bool {
	fields1 := searchSort.GetSort()
	return cindex.FIELD_DOC.Equals(fields1[0])
}

func canEarlyTerminateOnPrefix(searchSort, indexSort index.Sort) bool {
	if indexSort != nil {
		fields1 := searchSort.GetSort()
		fields2 := indexSort.GetSort()
		// early termination is possible if fields1 is a prefix of fields2
		if len(fields1) > len(fields2) {
			return false
		}

		for i, field := range fields1 {
			if !field.Equals(fields2[i]) {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

func (t *TopFieldCollector) updateBottom(doc int) {
	// bottom.score is already set to Float.NaN in add().
	t.bottom.doc = t.docBase + doc
	t.bottom = t.pq.UpdateTop()
}
