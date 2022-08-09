package index

import (
	"io"
)

// FieldsProducer Sole constructor. (For invocation by subclass constructors, typically implicit.)
type FieldsProducer interface {
	io.Closer

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum value against large
	// data files.
	CheckIntegrity() error

	// GetMergeInstance Returns an instance optimized for merging. This instance may only be consumed in the
	// thread that called getMergeInstance().
	// The default implementation returns this
	GetMergeInstance() FieldsProducer
}
