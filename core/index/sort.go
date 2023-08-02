package index

import "github.com/geange/lucene-go/core/interface/index"

type sortBase struct {
	fields []index.SortField
}

func NewSort(fields []index.SortField) index.Sort {
	return &sortBase{fields: fields}
}

// SetSort Sets the sort to the given criteria.
func (s *sortBase) SetSort(fields []index.SortField) {
	s.fields = fields
}

// GetSort Representation of the sort criteria.
// Returns: Array of SortField objects used in this sort criteria
func (s *sortBase) GetSort() []index.SortField {
	return s.fields
}
