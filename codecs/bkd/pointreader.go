package bkd

import "io"

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
