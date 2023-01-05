package index

// SegmentCommitInfo Embeds a [read-only] SegmentInfo and adds per-commit fields.
// lucene.experimental
type SegmentCommitInfo struct {
	// The SegmentInfo that we wrap.
	info *SegmentInfo

	// Id that uniquely identifies this segment commit.
	id []byte

	// How many deleted docs in the segment:
	delCount int

	// How many soft-deleted docs in the segment that are not also hard-deleted:
	softDelCount int

	// Generation number of the live docs file (-1 if there
	// are no deletes yet):
	delGen int64

	// Normally 1+delGen, unless an exception was hit on last
	// attempt to write:
	nextWriteDelGen int64

	// Generation number of the FieldInfos (-1 if there are no updates)
	fieldInfosGen int64

	// Normally 1+fieldInfosGen, unless an exception was hit on last attempt to
	// write
	nextWriteFieldInfosGen int64

	// Generation number of the DocValues (-1 if there are no updates)
	docValuesGen int64

	// Normally 1+dvGen, unless an exception was hit on last attempt to
	// write
	nextWriteDocValuesGen int64

	// Track the per-field DocValues update files
	dvUpdatesFiles map[int]map[string]struct{}

	// TODO should we add .files() to FieldInfosFormat, like we have on
	// LiveDocsFormat?
	// track the fieldInfos update files
	fieldInfosFiles map[string]struct{}

	sizeInBytes int64

	// NOTE: only used in-RAM by IW to track buffered deletes;
	// this is never written to/read from the Directory
	bufferedDeletesGen int64
}
