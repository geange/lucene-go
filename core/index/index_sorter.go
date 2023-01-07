package index

// IndexSorter Handles how documents should be sorted in an index, both within a segment and
// between segments. Implementers must provide the following methods:
// getDocComparator(LeafReader, int) - an object that determines how documents within a segment
// are to be sorted getComparableProviders(List) - an array of objects that return a sortable
// long value per document and segment getProviderName() - the SPI-registered name of a
// SortFieldProvider to serialize the sort The companion SortFieldProvider should be
// registered with SPI via META-INF/services
type IndexSorter interface {

	// GetComparableProviders Get an array of IndexSorter.ComparableProvider, one per segment,
	// for merge sorting documents in different segments
	// Params: readers – the readers to be merged
	GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error)

	// GetDocComparator Get a comparator that determines the sort order of docs within a single Reader.
	// NB We cannot simply use the FieldComparator API because it requires docIDs to be sent in-order.
	// The default implementations allocate array[maxDoc] to hold native values for comparison, but 1)
	// they are transient (only alive while sorting this one segment) and 2) in the typical index
	// sorting case, they are only used to sort newly flushed segments, which will be smaller than
	// merged segments
	//
	// Params: reader – the Reader to sort
	//		   maxDoc – the number of documents in the Reader
	GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error)

	// GetProviderName The SPI-registered name of a SortFieldProvider that will deserialize the parent SortField
	GetProviderName() string
}

// ComparableProvider Used for sorting documents across segments
type ComparableProvider interface {
	// GetAsComparableLong Returns a long so that the natural ordering of long values
	// matches the ordering of doc IDs for the given comparator
	GetAsComparableLong(docID int) (int64, error)
}

// DocComparator A comparator of doc IDs, used for sorting documents within a segment
type DocComparator interface {
	// Compare docID1 against docID2. The contract for the return value is the same as Comparator.compare(Object, Object).
	Compare(docID1, docID2 int) int
}
