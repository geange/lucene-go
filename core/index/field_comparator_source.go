package index

// FieldComparatorSource Provides a FieldComparator for custom field sorting.
// lucene.experimental
type FieldComparatorSource interface {
	NewComparator(fieldName string, numHits, sortPos int, reversed bool) FieldComparator
}
