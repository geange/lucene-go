package index

type Sort struct {
	fields []SortField
}

func NewSort(fields []SortField) *Sort {
	return &Sort{fields: fields}
}

// SetSort Sets the sort to the given criteria.
func (s *Sort) SetSort(fields []SortField) {
	s.fields = fields
}

// GetSort Representation of the sort criteria.
// Returns: Array of SortField objects used in this sort criteria
func (s *Sort) GetSort() []SortField {
	return s.fields
}
