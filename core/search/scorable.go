package search

import "github.com/geange/lucene-go/core/interface/search"

type BaseScorable struct {
}

func (*BaseScorable) SmoothingScore(docId int) (float64, error) {
	return 0, nil
}

func (*BaseScorable) SetMinCompetitiveScore(minScore float64) error {
	return nil
}

func (*BaseScorable) GetChildren() ([]search.ChildScorable, error) {
	return []search.ChildScorable{}, nil
}

// ChildScorable
// A child Scorer and its relationship to its parent. the meaning of the relationship
// depends upon the parent query.
type childScorable struct {

	// Child Scorer. (note this is typically a direct child, and may itself also have children).
	Child search.Scorable

	// An arbitrary string relating this scorer to the parent.
	Relationship string
}

func (c *childScorable) GetChild() search.Scorable {
	return c.Child
}

func (c *childScorable) GetRelationship() string {
	return c.Relationship
}

func NewChildScorable(child search.Scorable, relationship string) search.ChildScorable {
	return &childScorable{Child: child, Relationship: relationship}
}
