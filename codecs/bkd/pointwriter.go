package bkd

import "io"

// PointWriter Appends many points, and then at the end provides a PointReader to iterate those points.
// This abstracts away whether we write to disk, or use simple arrays in heap.
// lucene.internal
type PointWriter interface {
	io.Closer

	// Append Add a new point from the packed value and docId
	Append(packedValue []byte, docID int) error

	// AppendValue Add a new point from a PointValue
	AppendValue(pointValue PointValue) error

	// GetReader Returns a PointReader iterator to step through all previously added points
	GetReader(startPoint, length int64) (PointReader, error)

	// Count Return the number of points in this writer
	Count() int64

	// Destroy Removes any temp files behind this writer
	Destroy() error
}
