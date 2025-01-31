package search

import (
	"context"
	"io"
	"sort"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.Scorer = &BlockMaxConjunctionScorer{}

type BlockMaxConjunctionScorer struct {
	*BaseScorer

	scorers            []index.Scorer
	approximations     []types.DocIdSetIterator
	twoPhases          []index.TwoPhaseIterator
	maxScorePropagator *MaxScoreSumPropagator
	minScore           float64
}

func NewBlockMaxConjunctionScorer(weight index.Weight, scorersList []index.Scorer) (*BlockMaxConjunctionScorer, error) {
	res := &BlockMaxConjunctionScorer{
		BaseScorer:         NewScorer(weight),
		scorers:            scorersList,
		approximations:     make([]types.DocIdSetIterator, len(scorersList)),
		twoPhases:          nil,
		maxScorePropagator: nil,
		minScore:           0,
	}

	// Sort res by cost
	sort.Sort(sortScorerByCost(res.scorers))

	maxScorePropagator, err := NewMaxScoreSumPropagator(res.scorers)
	if err != nil {
		return nil, err
	}
	res.maxScorePropagator = maxScorePropagator

	twoPhaseList := make([]index.TwoPhaseIterator, 0)
	for i := range res.scorers {
		scorer := res.scorers[i]
		twoPhase := scorer.TwoPhaseIterator()
		if twoPhase != nil {
			twoPhaseList = append(twoPhaseList, twoPhase)
			res.approximations[i] = twoPhase.Approximation()
		} else {
			res.approximations[i] = scorer.Iterator()
		}
		if _, err := scorer.AdvanceShallow(0); err != nil {
			return nil, err
		}
	}
	res.twoPhases = twoPhaseList
	sort.Sort(sortTwoPhaseIteratorByMatchCost(res.twoPhases))
	return res, nil
}

var _ sort.Interface = &sortScorerByCost{}

type sortScorerByCost []index.Scorer

func (s sortScorerByCost) Len() int {
	return len(s)
}

func (s sortScorerByCost) Less(i, j int) bool {
	return s[i].Iterator().Cost() < s[j].Iterator().Cost()
}

func (s sortScorerByCost) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

var _ sort.Interface = &sortTwoPhaseIteratorByMatchCost{}

type sortTwoPhaseIteratorByMatchCost []index.TwoPhaseIterator

func (s sortTwoPhaseIteratorByMatchCost) Len() int {
	return len(s)
}

func (s sortTwoPhaseIteratorByMatchCost) Less(i, j int) bool {
	return s[i].MatchCost() < s[j].MatchCost()
}

func (s sortTwoPhaseIteratorByMatchCost) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (b *BlockMaxConjunctionScorer) Score() (float64, error) {
	score := 0.0
	for _, scorer := range b.scorers {
		num, err := scorer.Score()
		if err != nil {
			return 0, err
		}
		score += num
	}
	return score, nil
}

func (b *BlockMaxConjunctionScorer) DocID() int {
	return b.scorers[0].DocID()
}

func (b *BlockMaxConjunctionScorer) Iterator() types.DocIdSetIterator {
	if len(b.twoPhases) == 0 {
		return b.approximation()
	}
	return AsDocIdSetIterator(b.TwoPhaseIterator())
}

func (b *BlockMaxConjunctionScorer) GetMaxScore(upTo int) (float64, error) {
	sum := 0.0
	for _, scorer := range b.scorers {
		score, err := scorer.GetMaxScore(upTo)
		if err != nil {
			return 0, err
		}
		sum += score
	}
	return sum, nil
}

func (b *BlockMaxConjunctionScorer) SetMinCompetitiveScore(score float64) error {
	b.minScore = score
	return b.maxScorePropagator.SetMinCompetitiveScore(score)
}

func (b *BlockMaxConjunctionScorer) GetChildren() ([]index.ChildScorable, error) {
	children := make([]index.ChildScorable, 0)
	for _, scorer := range b.scorers {
		scorables, err := scorer.GetChildren()
		if err != nil {
			return nil, err
		}
		children = append(children, scorables...)
	}
	return children, nil
}

func (b *BlockMaxConjunctionScorer) TwoPhaseIterator() index.TwoPhaseIterator {
	if len(b.twoPhases) == 0 {
		return nil
	}

	cost := 0.0
	for _, phase := range b.twoPhases {
		cost += phase.MatchCost()
	}

	approx := b.approximation()

	return &bmcTwoPhaseIterator{
		approx:    approx,
		matchCost: cost,
	}
}

var _ index.TwoPhaseIterator = &bmcTwoPhaseIterator{}

type bmcTwoPhaseIterator struct {
	approx    types.DocIdSetIterator
	matchCost float64
	p         *BlockMaxConjunctionScorer
}

func (b *bmcTwoPhaseIterator) Approximation() types.DocIdSetIterator {
	return b.approx
}

func (b *bmcTwoPhaseIterator) Matches() (bool, error) {
	for _, twoPhase := range b.p.twoPhases {
		if ok, err := twoPhase.Matches(); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (b *bmcTwoPhaseIterator) MatchCost() float64 {
	return b.matchCost
}

func (b *BlockMaxConjunctionScorer) approximation() types.DocIdSetIterator {
	lead := b.approximations[0]
	return &bmcDocIdSetIterator{
		maxScore: 0,
		upTo:     -1,
		lead:     lead,
		p:        b,
	}
}

var _ types.DocIdSetIterator = &bmcDocIdSetIterator{}

type bmcDocIdSetIterator struct {
	maxScore float64
	upTo     int
	lead     types.DocIdSetIterator
	p        *BlockMaxConjunctionScorer
}

func (b *bmcDocIdSetIterator) DocID() int {
	return b.lead.DocID()
}

func (b *bmcDocIdSetIterator) NextDoc(ctx context.Context) (int, error) {
	return b.Advance(ctx, b.DocID()+1)
}

func (b *bmcDocIdSetIterator) Advance(ctx context.Context, target int) (int, error) {
	advanceTarget, err := b.advanceTarget(target)
	if err != nil {
		return 0, err
	}

	advance, err := b.lead.Advance(nil, advanceTarget)
	if err != nil {
		return 0, err
	}

	return b.doNext(ctx, advance)
}

func (b *bmcDocIdSetIterator) doNext(ctx context.Context, doc int) (int, error) {
advanceHead:
	for {
		if doc == types.NO_MORE_DOCS {
			return 0, io.EOF
		}

		if doc > b.upTo {
			// This check is useful when scorers return information about blocks
			// that do not actually have any matches. Otherwise `doc` will always
			// be in the current block already since it is always the result of
			// lead.advance(advanceTarget(some_doc_id))
			nextTarget, err := b.advanceTarget(doc)
			if err != nil {
				return 0, err
			}
			if nextTarget != doc {
				doc, err = b.lead.Advance(ctx, nextTarget)
				if err != nil {
					return 0, err
				}
				continue
			}
		}

		// then find agreement with other iterators
		for i := 0; i < len(b.p.approximations); i++ {
			other := b.p.approximations[i]

			// other.doc may already be equal to doc if we "continued advanceHead"
			// on the previous iteration and the advance on the lead scorer exactly matched.
			if other.DocID() < doc {
				next, err := other.Advance(ctx, doc)
				if err != nil {
					return 0, err
				}

				if next > doc {
					// iterator beyond the current doc - advance lead and continue to the new highest doc.
					advanceTarget, err := b.advanceTarget(next)
					if err != nil {
						return 0, err
					}
					doc, err = b.lead.Advance(ctx, advanceTarget)
					if err != nil {
						return 0, err
					}
					continue advanceHead
				}
			}
		}
		return doc, nil
	}
}

func (b *bmcDocIdSetIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, b, target)
}

func (b *bmcDocIdSetIterator) Cost() int64 {
	return b.lead.Cost()
}

func (b *bmcDocIdSetIterator) moveToNextBlock(target int) error {
	upTo, err := b.p.AdvanceShallow(target)
	if err != nil {
		return err
	}
	b.upTo = upTo

	maxScore, err := b.p.GetMaxScore(b.upTo)
	if err != nil {
		return err
	}
	b.maxScore = maxScore
	return nil
}

func (b *bmcDocIdSetIterator) advanceTarget(target int) (int, error) {
	if target > b.upTo {
		if err := b.moveToNextBlock(target); err != nil {
			return 0, err
		}
	}

	for {
		if b.maxScore >= b.p.minScore {
			return target, nil
		}

		target = b.upTo + 1

		if err := b.moveToNextBlock(target); err != nil {
			return 0, err
		}
	}
}
