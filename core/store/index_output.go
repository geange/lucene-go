package store

import "io"

// IndexOutput A DataOutput for appending data to a file in a Directory. Instances of this class are not thread-safe.
// See Also: Directory, IndexInput
type IndexOutput interface {
	io.Closer

	DataOutput

	// GetFilePointer Returns the current position in this file, where the next write will occur.
	GetFilePointer() int64
}
