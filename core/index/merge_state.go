package index

// MergeState Holds common state used during segment merging.
type MergeState struct {
	// Maps document IDs from old segments to document IDs in the new segment
	DocMaps []DocMap

	// SegmentInfo of the newly merged segment.
	segmentInfo *SegmentInfo
}

type DocMap interface {
	// Get Return the mapped docID or -1 if the given doc is not mapped.
	Get(docID int) int
}
