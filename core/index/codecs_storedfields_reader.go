package index

import (
	"github.com/geange/lucene-go/core/document"
	"io"
)

type StoredFieldsReader interface {
	io.Closer

	// VisitDocument Visit the stored fields for document docID
	VisitDocument(docID int, visitor document.StoredFieldVisitor) error

	Clone() StoredFieldsReader

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// value against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetMergeInstance Returns an instance optimized for merging. This instance may not be cloned.
	//The default implementation returns this
	GetMergeInstance() StoredFieldsReader
}
