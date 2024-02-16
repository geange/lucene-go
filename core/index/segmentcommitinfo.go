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

func NewSegmentCommitInfo(info *SegmentInfo, delCount, softDelCount int, delGen, fieldInfosGen, docValuesGen int64, id []byte) *SegmentCommitInfo {
	nextWriteDelGen := delGen + 1
	if delGen == -1 {
		nextWriteDelGen = 1
	}

	nextWriteFieldInfosGen := fieldInfosGen + 1
	if fieldInfosGen == -1 {
		nextWriteFieldInfosGen = 1
	}

	nextWriteDocValuesGen := docValuesGen + 1
	if docValuesGen == -1 {
		nextWriteDocValuesGen = 1
	}

	return &SegmentCommitInfo{
		info:                   info,
		id:                     id,
		delCount:               delCount,
		softDelCount:           softDelCount,
		delGen:                 delGen,
		nextWriteDelGen:        nextWriteDelGen,
		fieldInfosGen:          fieldInfosGen,
		nextWriteFieldInfosGen: nextWriteFieldInfosGen,
		docValuesGen:           docValuesGen,
		nextWriteDocValuesGen:  nextWriteDocValuesGen,
		dvUpdatesFiles:         map[int]map[string]struct{}{},
		fieldInfosFiles:        map[string]struct{}{},
		sizeInBytes:            0,
		bufferedDeletesGen:     0,
	}
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

func (s *SegmentCommitInfo) GetDelCountV1(includeSoftDeletes bool) int {
	if includeSoftDeletes {
		return s.GetDelCount() + s.GetSoftDelCount()
	}
	return s.GetDelCount()
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

func (s *SegmentCommitInfo) Files() (map[string]struct{}, error) {
	files := s.info.Files()

	// Must separately add any live docs files:
	_, err := s.info.GetCodec().LiveDocsFormat().Files(s, files)
	if err != nil {
		return nil, err
	}

	// must separately add any field updates files
	for _, updateFiles := range s.dvUpdatesFiles {
		for k := range updateFiles {
			files[k] = struct{}{}
		}
	}

	// must separately add fieldInfos files
	for k := range s.fieldInfosFiles {
		files[k] = struct{}{}
	}
	return files, nil
}

func (s *SegmentCommitInfo) GetNextWriteDelGen() int64 {
	return s.nextWriteDelGen
}

func (s *SegmentCommitInfo) SetNextWriteDelGen(v int64) {
	s.nextWriteDelGen = v
}

func (s *SegmentCommitInfo) GetNextWriteFieldInfosGen() int64 {
	return s.nextWriteFieldInfosGen
}

func (s *SegmentCommitInfo) SetNextWriteFieldInfosGen(v int64) {
	s.nextWriteFieldInfosGen = v
}

func (s *SegmentCommitInfo) GetNextWriteDocValuesGen() int64 {
	return s.nextWriteDocValuesGen
}

func (s *SegmentCommitInfo) SetNextWriteDocValuesGen(v int64) {
	s.nextWriteDocValuesGen = v
}

func (s *SegmentCommitInfo) SetFieldInfosFiles(fieldInfosFiles map[string]struct{}) {
	s.fieldInfosFiles = map[string]struct{}{}
	for file := range fieldInfosFiles {
		s.fieldInfosFiles[s.info.NamedForThisSegment(file)] = struct{}{}
	}
}

func (s *SegmentCommitInfo) SetDocValuesUpdatesFiles(files map[int]map[string]struct{}) {
	s.dvUpdatesFiles = map[int]map[string]struct{}{}
	for k, values := range files {
		newValues := make(map[string]struct{})
		for v := range values {
			newValues[v] = struct{}{}
		}
		s.dvUpdatesFiles[k] = newValues
	}
}

func (s *SegmentCommitInfo) Clone() *SegmentCommitInfo {
	other := NewSegmentCommitInfo(s.info, s.delCount, s.softDelCount, s.delGen, s.fieldInfosGen, s.docValuesGen, s.GetId())
	// Not clear that we need to carry over nextWriteDelGen
	// (i.e. do we ever clone after a failed write and
	// before the next successful write?), but just do it to
	// be safe:
	other.nextWriteDelGen = s.nextWriteDelGen
	other.nextWriteFieldInfosGen = s.nextWriteFieldInfosGen
	other.nextWriteDocValuesGen = s.nextWriteDocValuesGen

	for k, files := range s.dvUpdatesFiles {
		values := make(map[string]struct{}, len(files))
		for k := range files {
			values[k] = struct{}{}
		}
		other.dvUpdatesFiles[k] = values
	}

	for k := range s.fieldInfosFiles {
		other.fieldInfosFiles[k] = struct{}{}
	}

	return other
}

func (s *SegmentCommitInfo) GetId() []byte {
	if len(s.id) > 0 {
		items := make([]byte, len(s.id))
		copy(items, s.id)
		return items
	}
	return nil
}

func (s *SegmentCommitInfo) SizeInBytes() (int64, error) {
	if s.sizeInBytes == -1 {
		sum := int64(0)

		files, _ := s.Files()
		for fileName := range files {
			fileLength, err := s.info.dir.FileLength(nil, fileName)
			if err != nil {
				return 0, err
			}
			sum += fileLength
		}
		s.sizeInBytes = sum
	}

	return s.sizeInBytes, nil
}
