package search

import "github.com/geange/lucene-go/core/index"

var _ DisjunctionScorer = &DisjunctionSumScorer{}

// DisjunctionSumScorer
// A Scorer for OR like queries, counterpart of ConjunctionScorer.
type DisjunctionSumScorer struct {
}

func newDisjunctionScorer(weight Weight, subScorers []Scorer, scoreMode *ScoreMode) (*DisjunctionSumScorer, error) {
	panic("")
}

func (d *DisjunctionSumScorer) Score() (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) SmoothingScore(docId int) (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) SetMinCompetitiveScore(minScore float32) error {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) GetChildren() ([]ChildScorable, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) GetWeight() Weight {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) Iterator() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) TwoPhaseIterator() TwoPhaseIterator {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) AdvanceShallow(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) GetMaxScore(upTo int) (float32, error) {
	//TODO implement me
	panic("implement me")
}
