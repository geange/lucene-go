package index

import "io"

// TermVectorsReader Codec API for reading term vectors:
// lucene.experimental
type TermVectorsReader interface {
	io.Closer

	// Get Returns term vectors for this document, or null if term vectors were not indexed.
	// If offsets are available they are in an OffsetAttribute available from the
	// org.apache.lucene.index.PostingsEnum.
	Get(doc int) (Fields, error)

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// Clone Create a clone that one caller at a time may use to read term vectors.
	Clone() TermVectorsReader

	// GetMergeInstance Returns an instance optimized for merging. This instance may
	// only be consumed in the thread that called getMergeInstance().
	// The default implementation returns this
	GetMergeInstance() TermVectorsReader
}
