package index

import "github.com/geange/lucene-go/core/store"

// CompoundFormat Encodes/decodes compound files
// lucene.experimental
type CompoundFormat interface {

	// GetCompoundReader Returns a Directory view (read-only) for the compound files in this segment
	GetCompoundReader(dir store.Directory, si *SegmentInfo, context *store.IOContext) (CompoundDirectory, error)

	// Write Packs the provided segment's files into a compound format. All files referenced
	// by the provided SegmentInfo must have CodecUtil.writeIndexHeader and CodecUtil.writeFooter.
	Write(dir store.Directory, si *SegmentInfo, context *store.IOContext) error
}
