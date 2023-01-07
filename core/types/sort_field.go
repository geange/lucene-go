package types

// SortField Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortField struct {
	field        string
	stype        SortFieldType
	reverse      bool
	canUsePoints bool
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
