package index

import "github.com/bits-and-blooms/bitset"

// MergeState Holds common state used during segment merging.
type MergeState struct {
	// Maps document IDs from old segments to document IDs in the new segment
	DocMaps []DocMap

	// SegmentInfo of the newly merged segment.
	SegmentInfo *SegmentInfo

	// FieldInfos of the newly merged segment.
	MergeFieldInfos *FieldInfos

	// Stored field producers being merged
	StoredFieldsReaders []StoredFieldsReader

	// Term vector producers being merged
	TermVectorsReaders []TermVectorsReader

	// Norms producers being merged
	NormsProducers []NormsProducer

	// DocValues producers being merged
	DocValuesProducers []DocValuesProducer

	// FieldInfos being merged
	FieldInfos []FieldInfos

	// Live docs for each reader
	LiveDocs []*bitset.BitSet

	// Postings to merge
	FieldsProducers []FieldsProducer

	// Point readers to merge
	PointsReaders []PointsReader

	// Max docs per reader
	MaxDocs []int

	// InfoStream for debugging messages.

	// Indicates if the index needs to be sorted
	NeedsIndexSort bool
}
