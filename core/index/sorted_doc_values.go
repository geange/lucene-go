package index

import (
	"github.com/geange/lucene-go/core/util/automaton"
)

// SortedDocValues A per-document byte[] with presorted values. This is fundamentally an iterator over the
// int ord values per document, with random access APIs to resolve an int ord to BytesRef.
// Per-Document values in a SortedDocValues are deduplicated, dereferenced, and sorted into a dictionary of
// unique values. A pointer to the dictionary value (ordinal) can be retrieved for each document. Ordinals
// are dense and in increasing sorted order.
type SortedDocValues interface {
	SortedDocValuesExtra

	BinaryDocValues

	// OrdValue Returns the ordinal for the current docID. It is illegal to call this method after
	// advanceExact(int) returned false.
	// Returns: ordinal for the document: this is dense, starts at 0, then increments by 1 for the
	// next value in sorted order.
	OrdValue() (int, error)

	// LookupTerm If key exists, returns its ordinal, else returns -insertionPoint-1, like Arrays.binarySearch.
	// Params: key – Key to look up
	LookupTerm(key []byte) (int, error)

	// TermsEnum Returns a TermsEnum over the values. The enum supports TermsEnum.ord() and TermsEnum.seekExact(long).
	TermsEnum() (TermsEnum, error)

	// Intersect Returns a TermsEnum over the values, filtered by a CompiledAutomaton The enum supports TermsEnum.ord().
	Intersect(automaton *automaton.CompiledAutomaton) (TermsEnum, error)
}

type SortedDocValuesExtra interface {
	// LookupOrd Retrieves the value for the specified ordinal. The returned BytesRef may be re-used
	// across calls to lookupOrd(int) so make sure to copy it if you want to keep it around.
	// Params: ord – ordinal to lookup (must be >= 0 and < getValueCount())
	// See Also: ordValue()
	LookupOrd(ord int) ([]byte, error)

	// GetValueCount Returns the number of unique values.
	// Returns: number of unique values in this SortedDocValues. This is also equivalent to one plus the maximum ordinal.
	GetValueCount() int
}

type SortedDocValuesImp struct {
	extra SortedDocValuesExtra
}

func NewSortedDocValuesImp(extra SortedDocValuesExtra) *SortedDocValuesImp {
	return &SortedDocValuesImp{extra: extra}
}

func (r *SortedDocValuesImp) BinaryValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (r *SortedDocValuesImp) LookupTerm(key []byte) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *SortedDocValuesImp) TermsEnum() (TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (r *SortedDocValuesImp) Intersect(automaton *automaton.CompiledAutomaton) (TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}
