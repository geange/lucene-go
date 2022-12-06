package index

import "github.com/geange/lucene-go/core/store"

// SegmentReadState Holder class for common parameters used during read.
// lucene.experimental
type SegmentReadState struct {

	//Directory where this segment is read from.
	Directory store.Directory

	// SegmentInfo describing this segment.
	SegmentInfo *SegmentInfo

	// FieldInfos describing all fields in this segment.
	FieldInfos *FieldInfos

	// IOContext to pass to Directory.openInput(String, IOContext).
	Context *store.IOContext

	// Unique suffix for any postings files read for this segment.
	// PerFieldPostingsFormat sets this for each of the postings formats it wraps.
	// If you create a new PostingsFormat then any files you write/read must be derived
	// using this suffix (use IndexFileNames.segmentFileName(String, String, String)).
	SegmentSuffix string
}
