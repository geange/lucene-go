package index

import "github.com/geange/lucene-go/core/store"

// CompoundFormat Encodes/decodes compound files
// lucene.experimental
type CompoundFormat interface {

	// Returns a Directory view (read-only) for the compound files in this segment
	GetCompoundReader(dir store.Directory, si *SegmentInfo, context *store.IOContext) (CompoundDirectory, error)
}
