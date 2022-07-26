package search

// Explanation Expert: Describes the Score computation for document and query.
type Explanation struct {
	match       bool          // whether the document matched
	value       interface{}   // the value of this node
	description string        // what it represents
	details     []Explanation // sub-explanations
}

func NewExplanation(match bool, value interface{}, description string, details ...Explanation) *Explanation {
	return &Explanation{
		match:       match,
		value:       value,
		description: description,
		details:     details,
	}
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
func (e *Explanation) GetDetails() []Explanation {
	return e.details
}
