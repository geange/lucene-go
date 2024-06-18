package search

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

// ChildScorable
// A child Scorer and its relationship to its parent. the meaning of the relationship
// depends upon the parent query.
type childScorable struct {

	// Child Scorer. (note this is typically a direct child, and may itself also have children).
	Child Scorable

	// An arbitrary string relating this scorer to the parent.
	Relationship string
}

func (c *childScorable) GetChild() Scorable {
	return c.Child
}

func (c *childScorable) GetRelationship() string {
	return c.Relationship
}

func NewChildScorable(child Scorable, relationship string) ChildScorable {
	return &childScorable{Child: child, Relationship: relationship}
}
