package index

import (
	"github.com/geange/lucene-go/core/store"
)

// The SegmentMerger class combines two or more Segments, represented by an IndexReader,
// into a single Segment. Call the merge method to combine the segments.
type SegmentMerger struct {
	directory         store.Directory
	codec             Codec
	mergeState        *MergeState
	fieldInfosBuilder *FieldInfosBuilder
}

func NewSegmentMerger(readers []CodecReader, segmentInfo *SegmentInfo, dir store.Directory,
	fieldNumbers *FieldNumbers, ioCtx *store.IOContext) (*SegmentMerger, error) {

	//if ioCtx.Type != store.CONTEXT_MERGE {
	//	return nil, errors.New("context type should be MERGE")
	//}
	//
	//mergeState := store.NewMer
	// TODO: fix it
	panic("")
}

func (s *SegmentMerger) ShouldMerge() bool {
	maxDoc, _ := s.mergeState.SegmentInfo.MaxDoc()
	return maxDoc > 0
}
