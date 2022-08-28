package store

import "io"

// IndexOutput A DataOutput for appending data to a file in a Directory. Instances of this class are not thread-safe.
// See Also: Directory, IndexInput
type IndexOutput interface {
	io.Closer

	DataOutput

	// GetFilePointer Returns the current position in this file, where the next write will occur.
	GetFilePointer() int64

	GetChecksum() (uint32, error)
}

type IndexOutputImp struct {
	*DataOutputImp

	// Just the name part from resourceDescription
	name string
}

func NewIndexOutputImp(output DataOutput, name string) *IndexOutputImp {
	return &IndexOutputImp{
		DataOutputImp: NewDataOutputImp(output),
		name:          name,
	}
}

func (i *IndexOutputImp) GetName() string {
	return i.name
}
