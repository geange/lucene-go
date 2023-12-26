package search

// Scorable
// Allows access to the Score of a Query
// 允许访问查询的分数
type Scorable interface {
	// Score Returns the Score of the current document matching the query.
	Score() (float64, error)

	// SmoothingScore Returns the smoothing Score of the current document matching the query. This Score
	// is used when the query/term does not appear in the document, and behaves like an idf. The smoothing
	// Score is particularly important when the Scorer returns a product of probabilities so that the
	// document Score does not go to zero when one probability is zero. This can return 0 or a smoothing Score.
	//
	// Smoothing scores are described in many papers, including: Metzler, D. and Croft, W. B. , "Combining
	// the Language Model and Inference Network Approaches to Retrieval," Information Processing and Management
	// Special Issue on Bayesian Networks and Information Retrieval, 40(5), pp.735-750.
	SmoothingScore(docId int) (float64, error)

	// DocID Returns the doc ID that is currently being scored.
	DocID() int

	// SetMinCompetitiveScore Optional method: Tell the scorer that its iterator may safely ignore all
	// documents whose Score is less than the given minScore. This is a no-op by default. This method
	// may only be called from collectors that use ScoreMode.TOP_SCORES, and successive calls may
	// only set increasing values of minScore.
	SetMinCompetitiveScore(minScore float64) error

	// GetChildren Returns child sub-scorers positioned on the current document
	GetChildren() ([]ChildScorable, error)
}

type BaseScorable struct {
}

func (*BaseScorable) SmoothingScore(docId int) (float64, error) {
	return 0, nil
}

func (*BaseScorable) SetMinCompetitiveScore(minScore float64) error {
	return nil
}

func (*BaseScorable) GetChildren() ([]ChildScorable, error) {
	return []ChildScorable{}, nil
}

// ChildScorable A child Scorer and its relationship to its parent. the meaning of the relationship
// depends upon the parent query.
type ChildScorable struct {

	// Child Scorer. (note this is typically a direct child, and may itself also have children).
	Child Scorable

	// An arbitrary string relating this scorer to the parent.
	Relationship string
}

func NewChildScorable(child Scorable, relationship string) *ChildScorable {
	return &ChildScorable{Child: child, Relationship: relationship}
}
