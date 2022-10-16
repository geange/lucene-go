package fst

import "github.com/geange/lucene-go/util/fst"

type INPUT_TYPE int

const (
	BYTE1 = INPUT_TYPE(iota)
	BYTE2
	BYTE4
)

type FST struct {
	inputType INPUT_TYPE

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput any

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB). If the FST is less than 1 GB then bytesArray is set instead.
	bytes BytesStore

	fstStore fst.FSTStore

	startNode int64

	outputs Outputs
}
