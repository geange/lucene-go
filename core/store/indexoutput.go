package store

import "io"

// IndexOutput A DataOutput for appending data to a file in a Directory. Instances of this class are not thread-safe.
// See Also: Directory, IndexInput
type IndexOutput interface {
	io.Closer

	DataOutput

	GetName() string

	// GetFilePointer Returns the current pos in this file,
	// where the next write will occur.
	GetFilePointer() int64

	GetChecksum() (uint32, error)
}

type IndexOutputBase struct {
	*Writer

	name string
}

func NewIndexOutputBase(name string, writer io.Writer) *IndexOutputBase {
	return &IndexOutputBase{
		Writer: NewWriter(writer),
		name:   name,
	}
}

func (r *IndexOutputBase) GetName() string {
	return r.name
}
