package index

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

// SegmentWriteState
// Holder class for common parameters used during write.
type SegmentWriteState struct {

	// Directory
	// where this segment will be written to.
	Directory store.Directory

	// SegmentInfo
	// describing this segment.
	SegmentInfo *SegmentInfo

	// FieldInfos
	// describing all fields in this segment.
	FieldInfos index.FieldInfos

	// DelCountOnFlush
	// Number of deleted documents set while flushing the segment.
	DelCountOnFlush int

	// SoftDelCountOnFlush
	// Number of only soft deleted documents set while flushing the segment.
	SoftDelCountOnFlush int

	// SegUpdates
	// Deletes and updates to apply while we are flushing the segment. A Term is enrolled in here if
	// it was deleted/updated at one point, and it's mapped to the docIDUpto, meaning any docID < docIDUpto
	// containing this term should be deleted/updated.
	SegUpdates *BufferedUpdates

	// LiveDocs
	// FixedBitSet recording live documents; this is only set if there is one or more deleted documents.
	LiveDocs *bitset.BitSet

	// SegmentSuffix
	// Unique suffix for any postings files written for this segment.
	// PerFieldPostingsFormat sets this for each of the postings formats it wraps.
	// If you create a new PostingsFormat then any files you write/read must be
	// derived using this suffix (use IndexFileNames.segmentFileName(String, String, String)).
	// Note: the suffix must be either empty, or be a textual suffix contain exactly two parts
	// (separated by underscore), or be a base36 generation.
	SegmentSuffix string

	// Context
	// IOContext for all writes; you should pass this to Directory.createOutput(String, IOContext).
	Context *store.IOContext
}

func NewSegmentWriteState(directory store.Directory, segmentInfo *SegmentInfo, fieldInfos index.FieldInfos, segUpdates *BufferedUpdates, ioContext *store.IOContext) *SegmentWriteState {

	return &SegmentWriteState{
		Directory:           directory,
		SegmentInfo:         segmentInfo,
		FieldInfos:          fieldInfos,
		DelCountOnFlush:     0,
		SoftDelCountOnFlush: 0,
		SegUpdates:          segUpdates,
		LiveDocs:            nil,
		SegmentSuffix:       "",
	}
}

// SegmentReadState
// Holder class for common parameters used during read.
// lucene.experimental
type SegmentReadState struct {

	//Directory where this segment is read from.
	Directory store.Directory

	// SegmentInfo describing this segment.
	SegmentInfo *SegmentInfo

	// FieldInfos describing all fields in this segment.
	FieldInfos index.FieldInfos

	// IOContext to pass to Directory.openInput(String, IOContext).
	Context *store.IOContext

	// Unique suffix for any postings files read for this segment.
	// PerFieldPostingsFormat sets this for each of the postings formats it wraps.
	// If you create a new PostingsFormat then any files you write/read must be derived
	// using this suffix (use IndexFileNames.segmentFileName(String, String, String)).
	SegmentSuffix string
}

func NewSegmentReadState(dir store.Directory, info *SegmentInfo,
	fieldInfos index.FieldInfos, ioContext *store.IOContext, segmentSuffix string) *SegmentReadState {

	return &SegmentReadState{
		Directory:     dir,
		SegmentInfo:   info,
		FieldInfos:    fieldInfos,
		Context:       ioContext,
		SegmentSuffix: segmentSuffix,
	}
}
