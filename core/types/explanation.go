package types

// Explanation
// Expert: Describes the score computation for document and query.
// 用于分数计算过程的描述
type Explanation struct {
	match       bool           // whether the document matched
	value       any            // the value of this node
	description string         // what it represents
	details     []*Explanation // sub-explanations
}

func NewExplanation(match bool, value any, description string, details ...*Explanation) *Explanation {
	return &Explanation{
		match:       match,
		value:       value,
		description: description,
		details:     details,
	}
}

// ExplanationMatch
// Create a new explanation for a match.
// Params: 	value – the contribution to the score of the document
//
//	description – how value was computed
//	details – sub explanations that contributed to this explanation
func ExplanationMatch(value any, description string, details ...*Explanation) *Explanation {
	return NewExplanation(true, value, description, details...)
}

// ExplanationNoMatch
// Create a new explanation for a document which does not match.
func ExplanationNoMatch(description string, details ...*Explanation) *Explanation {
	return NewExplanation(false, float64(0), description, details...)
}

// IsMatch Indicates whether or not this Explanation models a match.
func (e *Explanation) IsMatch() bool {
	return e.match
}

// GetValue The value assigned to this explanation node.
func (e *Explanation) GetValue() interface{} {
	return e.value
}

// GetDescription A description of this explanation node.
func (e *Explanation) GetDescription() string {
	return e.description
}

// GetDetails The sub-nodes of this explanation node.
func (e *Explanation) GetDetails() []*Explanation {
	return e.details
}
