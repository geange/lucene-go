package search

import "github.com/geange/lucene-go/core/index"

// TermScorer Expert: A Scorer for documents matching a Term.
type TermScorer struct {
	weight       Weight
	postingsEnum index.PostingsEnum
	impactsEnum  index.ImpactsEnum
	iterator     index.DocIdSetIterator
	docScorer    *LeafSimScorer
	impactsDisi  *ImpactsDISI
}

func NewTermScorerWithPostings(weight Weight, postingsEnum index.PostingsEnum, docScorer *LeafSimScorer) *TermScorer {
	this := &TermScorer{
		weight:       weight,
		iterator:     postingsEnum,
		postingsEnum: postingsEnum,
		impactsEnum:  index.NewSlowImpactsEnum(postingsEnum),
		docScorer:    docScorer,
	}

	this.impactsDisi = NewImpactsDISI(this.impactsEnum, this.impactsEnum, docScorer.GetSimScorer())
	return this
}

func NewTermScorerWithImpacts(weight Weight, impactsEnum index.ImpactsEnum, docScorer *LeafSimScorer) *TermScorer {
	this := &TermScorer{
		weight:       weight,
		postingsEnum: impactsEnum,
		impactsEnum:  impactsEnum,
		docScorer:    docScorer,
	}

	this.impactsDisi = NewImpactsDISI(this.impactsEnum, this.impactsEnum, docScorer.GetSimScorer())
	this.iterator = this.impactsDisi
	return this
}

func (t *TermScorer) Score() (float64, error) {
	freq, err := t.postingsEnum.Freq()
	if err != nil {
		return 0, err
	}
	return t.docScorer.Score(t.postingsEnum.DocID(), float64(freq))
}

func (t *TermScorer) SmoothingScore(docId int) (float64, error) {
	return t.docScorer.Score(docId, 0)
}

func (t *TermScorer) DocID() int {
	return t.postingsEnum.DocID()
}

func (t *TermScorer) Freq() (int, error) {
	return t.postingsEnum.Freq()
}

func (t *TermScorer) SetMinCompetitiveScore(minScore float64) error {
	return t.impactsDisi.setMinCompetitiveScore(minScore)
}

func (t *TermScorer) GetChildren() ([]ChildScorable, error) {
	return []ChildScorable{}, nil
}

func (t *TermScorer) GetWeight() Weight {
	return t.weight
}

func (t *TermScorer) Iterator() index.DocIdSetIterator {
	return t.iterator
}

func (t *TermScorer) TwoPhaseIterator() TwoPhaseIterator {
	return nil
}

func (t *TermScorer) GetMaxScore(upTo int) (float64, error) {
	return t.impactsDisi.getMaxScore(upTo)
}
