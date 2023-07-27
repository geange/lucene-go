package index

import (
	"io"

	"github.com/geange/lucene-go/core/document"
)

// NormsProducer Abstract API that produces field normalization values
type NormsProducer interface {
	io.Closer

	// GetNorms Returns NumericDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetNorms(field *document.FieldInfo) (NumericDocValues, error)

	// CheckIntegrity Checks consistency of this producer
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum item
	// against large data files.
	CheckIntegrity() error

	// GetMergeInstance Returns an instance optimized for merging. This instance may only be used from the
	// thread that acquires it.
	// The default implementation returns this
	GetMergeInstance() NormsProducer
}
