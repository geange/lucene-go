package compressing

import "io"

type FieldsIndex interface {
	io.Closer

	// GetStartPointer
	// Get the start pointer for the block that contains the given docID.
	GetStartPointer(docId int) int64

	// CheckIntegrity
	// Check the integrity of the index.
	CheckIntegrity() error

	Clone() FieldsIndex
}
