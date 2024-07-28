package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

type BaseScorable struct {
}

func (*BaseScorable) SmoothingScore(docId int) (float64, error) {
	return 0, nil
}

func (*BaseScorable) SetMinCompetitiveScore(minScore float64) error {
	return nil
}

func (*BaseScorable) GetChildren() ([]index.ChildScorable, error) {
	return []index.ChildScorable{}, nil
}

// ChildScorable
// A child Scorer and its relationship to its parent. the meaning of the relationship
// depends upon the parent query.
type childScorable struct {

	// Child Scorer. (note this is typically a direct child, and may itself also have children).
	Child index.Scorable

	// An arbitrary string relating this scorer to the parent.
	Relationship string
}

func (c *childScorable) GetChild() index.Scorable {
	return c.Child
}

func (c *childScorable) GetRelationship() string {
	return c.Relationship
}

func NewChildScorable(child index.Scorable, relationship string) index.ChildScorable {
	return &childScorable{Child: child, Relationship: relationship}
}
