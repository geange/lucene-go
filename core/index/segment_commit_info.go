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

func (s *SegmentCommitInfo) Info() *SegmentInfo {
	return s.info
}

// HasDeletions Returns true if there are any deletions for the segment at this commit.
func (s *SegmentCommitInfo) HasDeletions() bool {
	return s.delGen != -1
}

// HasFieldUpdates Returns true if there are any field updates for the segment in this commit.
func (s *SegmentCommitInfo) HasFieldUpdates() bool {
	return s.fieldInfosGen != -1
}

// GetNextFieldInfosGen Returns the next available generation number of the FieldInfos files.
func (s *SegmentCommitInfo) GetNextFieldInfosGen() int64 {
	return s.nextWriteFieldInfosGen
}

// GetFieldInfosGen Returns the generation number of the field infos file or -1 if there are no field updates yet.
func (s *SegmentCommitInfo) GetFieldInfosGen() int64 {
	return s.fieldInfosGen
}

// GetNextDocValuesGen Returns the next available generation number of the DocValues files.
func (s *SegmentCommitInfo) GetNextDocValuesGen() int64 {
	return s.nextWriteDocValuesGen
}

// GetDocValuesGen Returns the generation number of the DocValues file or -1 if there are no doc-values updates yet.
func (s *SegmentCommitInfo) GetDocValuesGen() int64 {
	return s.docValuesGen
}

// GetNextDelGen Returns the next available generation number of the live docs file.
func (s *SegmentCommitInfo) GetNextDelGen() int64 {
	return s.nextWriteDelGen
}

// GetDelGen Returns generation number of the live docs file or -1 if there are no deletes yet.
func (s *SegmentCommitInfo) GetDelGen() int64 {
	return s.delGen
}

// GetDelCount Returns the number of deleted docs in the segment.
func (s *SegmentCommitInfo) GetDelCount() int {
	return s.delCount
}

// GetSoftDelCount Returns the number of only soft-deleted docs.
func (s *SegmentCommitInfo) GetSoftDelCount() int {
	return s.softDelCount
}

func (s *SegmentCommitInfo) SetDelCount(delCount int) {
	s.delCount = delCount
}

func (s *SegmentCommitInfo) SetSoftDelCount(softDelCount int) {
	s.softDelCount = softDelCount
}
