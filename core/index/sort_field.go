package index

// SortField Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortField struct {
	field        string
	_type        SortFieldType
	reverse      bool
	canUsePoints bool
	missingValue any
}

// GetField Returns the name of the field. Could return null if the sort is by SCORE or DOC.
// Returns: Name of field, possibly null.
func (s *SortField) GetField() string {
	return s.field
}

// GetType Returns the type of contents in the field.
// Returns: One of the constants SCORE, DOC, STRING, INT or FLOAT.
func (s *SortField) GetType() SortFieldType {
	return s._type
}

// SetCanUsePoints For numeric sort fields, setting this field, indicates that the same numeric
// data has been indexed with two fields: doc values and points and that these fields have the
// same name. This allows to use sort optimization and skip non-competitive documents.
func (s *SortField) SetCanUsePoints() {
	s.canUsePoints = true
}

func (s *SortField) GetCanUsePoints() bool {
	return s.canUsePoints
}

// NeedsScores Whether the relevance score is needed to sort documents.
func (s *SortField) NeedsScores() bool {
	return s._type == SCORE
}

const (
	ProviderName = "SortField"
)

// GetIndexSorter Returns an IndexSorter used for sorting index segments by this SortField.
// If the SortField cannot be used for index sorting (for example, if it uses scores or other
// query-dependent values) then this method should return null SortFields that implement
// this method should also implement a companion SortFieldProvider to serialize and deserialize
// the sort in index segment headers
// lucene.experimental
func (s *SortField) GetIndexSorter() IndexSorter {
	switch s._type {
	case STRING:
		return nil
	case INT:
		return NewIntSorter(ProviderName, s.missingValue.(int32), s.reverse, func(reader LeafReader) (NumericDocValues, error) {
			return nil, nil
		})
	case LONG:
		return NewLongSorter(ProviderName, s.missingValue.(int64), s.reverse, func(reader LeafReader) (NumericDocValues, error) {
			return nil, nil
		})
	case DOUBLE:
		return nil
	case FLOAT:
		return nil
	default:
		return nil
	}
}

// SortFieldType Specifies the type of the terms to be sorted, or special types such as CUSTOM
type SortFieldType int

const (
	// SCORE // Sort by document score (relevance).
	// Sort values are Float and higher values are at the front.
	SCORE = SortFieldType(iota)

	// DOC Sort by document number (index order).
	// Sort values are Integer and lower values are at the front.
	DOC

	// STRING Sort using term values as Strings.
	// Sort values are String and lower values are at the front.
	STRING

	// INT Sort using term values as encoded Integers.
	// Sort values are Integer and lower values are at the front.
	INT

	// FLOAT Sort using term values as encoded Floats.
	// Sort values are Float and lower values are at the front.
	FLOAT

	// LONG Sort using term values as encoded Longs.
	// Sort values are Long and lower values are at the front.
	LONG

	// DOUBLE Sort using term values as encoded Doubles.
	// Sort values are Double and lower values are at the front.
	DOUBLE

	// CUSTOM Sort using a custom Comparator.
	// Sort values are any Comparable and sorting is done according to natural order.
	CUSTOM

	// STRING_VAL Sort using term values as Strings,
	// but comparing by value (using String.compareTo) for all comparisons.
	// This is typically slower than STRING, which uses ordinals to do the sorting.
	STRING_VAL

	// REWRITEABLE Force rewriting of SortField using rewrite(IndexSearcher) before it can be used for sorting
	REWRITEABLE
)
