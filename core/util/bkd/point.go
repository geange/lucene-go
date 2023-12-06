package bkd

import (
	"context"
	"io"
)

// PointValue Represents a dimensional point value written in the BKD tree.
// lucene.internal
type PointValue interface {

	// PackedValue Returns the packed values for the dimensions
	PackedValue() []byte

	// DocID Returns the docID
	DocID() int

	// PackedValueDocIDBytes Returns the byte representation of the packed value together with the docID
	PackedValueDocIDBytes() []byte
}

// PointReader One pass iterator through all points previously written with a PointWriter,
// abstracting away whether points are read from (offline) disk or simple arrays in heap.
// lucene.internal
type PointReader interface {
	io.Closer

	// Next Returns false once iteration is done, else true.
	Next() (bool, error)

	// PointValue Sets the packed value in the provided ByteRef
	PointValue() PointValue
}

// PointWriter Appends many points, and then at the end provides a PointReader to iterate those points.
// This abstracts away whether we write to disk, or use simple arrays in heap.
// lucene.internal
type PointWriter interface {
	io.Closer

	// Append Add a new point from the packed value and docId
	Append(ctx context.Context, packedValue []byte, docID int) error

	// AppendPoint Add a new point from a PointValue
	AppendPoint(pointValue PointValue) error

	// GetReader Returns a PointReader iterator to step through all previously added points
	GetReader(startPoint, length int) (PointReader, error)

	// Count Return the number of points in this writer
	Count() int

	// Destroy Removes any temp files behind this writer
	Destroy() error
}
