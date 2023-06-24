package search

import (
	"github.com/geange/lucene-go/core/index"
	"io"
	"sort"
)

var _ Scorer = &BlockMaxConjunctionScorer{}

type BlockMaxConjunctionScorer struct {
	*ScorerDefault

	scorers            []Scorer
	approximations     []index.DocIdSetIterator
	twoPhases          []TwoPhaseIterator
	maxScorePropagator *MaxScoreSumPropagator
	minScore           float64
}

func NewBlockMaxConjunctionScorer(weight Weight, scorersList []Scorer) (*BlockMaxConjunctionScorer, error) {
	res := &BlockMaxConjunctionScorer{
		ScorerDefault:      NewScorer(weight),
		scorers:            scorersList,
		approximations:     make([]index.DocIdSetIterator, len(scorersList)),
		twoPhases:          nil,
		maxScorePropagator: nil,
		minScore:           0,
	}

	// Sort res by cost
	sort.Sort(sortScorerByCost(res.scorers))
	var err error
	res.maxScorePropagator, err = NewMaxScoreSumPropagator(res.scorers)
	if err != nil {
		return nil, err
	}

	twoPhaseList := make([]TwoPhaseIterator, 0)
	for i := range res.scorers {
		scorer := res.scorers[i]
		twoPhase := scorer.TwoPhaseIterator()
		if twoPhase != nil {
			twoPhaseList = append(twoPhaseList, twoPhase)
			res.approximations[i] = twoPhase.Approximation()
		} else {
			res.approximations[i] = scorer.Iterator()
		}
		_, err := scorer.AdvanceShallow(0)
		if err != nil {
			return nil, err
		}
	}
	res.twoPhases = twoPhaseList
	sort.Sort(sortTwoPhaseIteratorByMatchCost(res.twoPhases))
	return res, nil
}

var _ sort.Interface = &sortScorerByCost{}

type sortScorerByCost []Scorer

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

type sortTwoPhaseIteratorByMatchCost []TwoPhaseIterator

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

func (b *BlockMaxConjunctionScorer) Iterator() index.DocIdSetIterator {
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

func (b *BlockMaxConjunctionScorer) GetChildren() ([]ChildScorable, error) {
	children := make([]ChildScorable, 0)
	for _, scorer := range b.scorers {
		scorables, err := scorer.GetChildren()
		if err != nil {
			return nil, err
		}
		children = append(children, scorables...)
	}
	return children, nil
}

func (b *BlockMaxConjunctionScorer) TwoPhaseIterator() TwoPhaseIterator {
	if len(b.twoPhases) == 0 {
		return nil
	}

	matchCost := 0.0
	for _, phase := range b.twoPhases {
		matchCost += phase.MatchCost()
	}

	approx := b.approximation()

	return &bmcTwoPhaseIterator{
		approx:    approx,
		matchCost: matchCost,
	}
}

var _ TwoPhaseIterator = &bmcTwoPhaseIterator{}

type bmcTwoPhaseIterator struct {
	approx    index.DocIdSetIterator
	matchCost float64
	p         *BlockMaxConjunctionScorer
}

func (b *bmcTwoPhaseIterator) Approximation() index.DocIdSetIterator {
	return b.approx
}

func (b *bmcTwoPhaseIterator) Matches() (bool, error) {
	for _, twoPhase := range b.p.twoPhases {
		if ok, _ := twoPhase.Matches(); !ok {
			return false, nil
		}
	}
	return true, nil
}

func (b *bmcTwoPhaseIterator) MatchCost() float64 {
	return b.matchCost
}

func (b *BlockMaxConjunctionScorer) approximation() index.DocIdSetIterator {
	lead := b.approximations[0]
	return &bmcDocIdSetIterator{
		maxScore: 0,
		upTo:     -1,
		lead:     lead,
		p:        b,
	}
}

var _ index.DocIdSetIterator = &bmcDocIdSetIterator{}

type bmcDocIdSetIterator struct {
	maxScore float64
	upTo     int
	lead     index.DocIdSetIterator
	p        *BlockMaxConjunctionScorer
}

func (b *bmcDocIdSetIterator) DocID() int {
	return b.lead.DocID()
}

func (b *bmcDocIdSetIterator) NextDoc() (int, error) {
	return b.Advance(b.DocID() + 1)
}

func (b *bmcDocIdSetIterator) Advance(target int) (int, error) {
	advanceTarget, err := b.advanceTarget(target)
	if err != nil {
		return 0, err
	}

	advance, err := b.lead.Advance(advanceTarget)
	if err != nil {
		return 0, err
	}

	return b.doNext(advance)
}

func (b *bmcDocIdSetIterator) doNext(doc int) (int, error) {
	for {
		if doc == index.NO_MORE_DOCS {
			return index.NO_MORE_DOCS, io.EOF
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
				doc, err = b.lead.Advance(nextTarget)
				if err != nil {
					return 0, err
				}
				continue
			}
		}
	}
}

func (b *bmcDocIdSetIterator) SlowAdvance(target int) (int, error) {
	return index.SlowAdvance(b, target)
}

func (b *bmcDocIdSetIterator) Cost() int64 {
	return b.lead.Cost()
}

func (b *bmcDocIdSetIterator) moveToNextBlock(target int) (err error) {
	b.upTo, err = b.SlowAdvance(target)
	if err != nil {
		return err
	}
	maxScore, err := b.p.GetMaxScore(b.upTo)
	if err != nil {
		return err
	}
	b.maxScore = maxScore
	return nil
}

func (b *bmcDocIdSetIterator) advanceTarget(target int) (int, error) {
	if target > b.upTo {
		err := b.moveToNextBlock(target)
		if err != nil {
			return 0, err
		}
	}

	for {
		if b.maxScore >= b.p.minScore {
			return target, nil
		}

		target = b.upTo + 1

		err := b.moveToNextBlock(target)
		if err != nil {
			return 0, err
		}
	}
}
