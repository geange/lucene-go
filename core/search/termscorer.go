package search

import (
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.Scorer = &TermScorer{}

// TermScorer
// Expert: A Scorer for documents matching a Term.
type TermScorer struct {
	*BaseScorer

	//weight       Weight
	postingsEnum index.PostingsEnum
	impactsEnum  index.ImpactsEnum
	iterator     types.DocIdSetIterator
	docScorer    *LeafSimScorer
	impactsDISI  *ImpactsDISI
}

func NewTermScorerWithPostings(weight index.Weight, postingsEnum index.PostingsEnum, docScorer *LeafSimScorer) *TermScorer {
	this := &TermScorer{
		iterator:     postingsEnum,
		postingsEnum: postingsEnum,
		impactsEnum:  coreIndex.NewSlowImpactsEnum(postingsEnum),
		docScorer:    docScorer,
	}
	this.BaseScorer = NewScorer(weight)

	this.impactsDISI = NewImpactsDISI(this.impactsEnum, this.impactsEnum, docScorer.GetSimScorer())
	return this
}

func NewTermScorerWithImpacts(weight index.Weight, impactsEnum index.ImpactsEnum, docScorer *LeafSimScorer) *TermScorer {
	termScorer := &TermScorer{
		postingsEnum: impactsEnum,
		impactsEnum:  impactsEnum,
		docScorer:    docScorer,
	}
	termScorer.BaseScorer = NewScorer(weight)

	termScorer.impactsDISI = NewImpactsDISI(termScorer.impactsEnum, termScorer.impactsEnum, docScorer.GetSimScorer())
	termScorer.iterator = termScorer.impactsDISI
	return termScorer
}

func (t *TermScorer) Score() (float64, error) {
	freq, err := t.postingsEnum.Freq()
	if err != nil {
		return 0, err
	}

	score, err := t.docScorer.Score(t.postingsEnum.DocID(), float64(freq))
	if err != nil {
		return 0, err
	}
	return score, nil
}

func (t *TermScorer) SmoothingScore(docId int) (float64, error) {
	score, err := t.docScorer.Score(docId, 0)
	if err != nil {
		return 0, err
	}
	return score, nil
}

func (t *TermScorer) DocID() int {
	return t.postingsEnum.DocID()
}

func (t *TermScorer) Freq() (int, error) {
	return t.postingsEnum.Freq()
}

func (t *TermScorer) SetMinCompetitiveScore(minScore float64) error {
	return t.impactsDISI.setMinCompetitiveScore(float64(minScore))
}

func (t *TermScorer) GetChildren() ([]index.ChildScorable, error) {
	return []index.ChildScorable{}, nil
}

func (t *TermScorer) GetWeight() index.Weight {
	return t.weight
}

func (t *TermScorer) Iterator() types.DocIdSetIterator {
	return t.iterator
}

func (t *TermScorer) TwoPhaseIterator() index.TwoPhaseIterator {
	return nil
}

func (t *TermScorer) GetMaxScore(upTo int) (float64, error) {
	score, err := t.impactsDISI.GetMaxScore(upTo)
	if err != nil {
		return 0, err
	}
	return score, nil
}
