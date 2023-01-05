package index

import "github.com/geange/lucene-go/core/store"

// TermVectorsFormat Controls the format of term vectors
type TermVectorsFormat interface {

	// VectorsReader Returns a TermVectorsReader to read term vectors.
	VectorsReader(dir store.Directory, segmentInfo *SegmentInfo,
		fieldInfos *FieldInfos, context *store.IOContext) (TermVectorsReader, error)

	// VectorsWriter Returns a TermVectorsWriter to write term vectors.
	VectorsWriter(dir store.Directory,
		segmentInfo *SegmentInfo, context *store.IOContext) (TermVectorsWriter, error)
}
