package index

import (
	"context"
	"maps"
	"slices"

	"github.com/google/uuid"
)

type SegmentCommitInfo interface {
	Info() SegmentInfo
	HasDeletions() bool
	HasFieldUpdates() bool
	GetNextFieldInfosGen() int64
	GetFieldInfosGen() int64
	GetNextDocValuesGen() int64
	GetDocValuesGen() int64
	GetNextDelGen() int64
	GetDelGen() int64
	GetDelCount() int
	GetDelCountWithSoftDeletes(includeSoftDeletes bool) int
	GetSoftDelCount() int
	SetDelCount(delCount int)
	SetSoftDelCount(softDelCount int)
	Files() (map[string]struct{}, error)
	GetNextWriteDelGen() int64
	SetNextWriteDelGen(v int64)
	GetNextWriteFieldInfosGen() int64
	SetNextWriteFieldInfosGen(v int64)
	GetNextWriteDocValuesGen() int64
	SetNextWriteDocValuesGen(v int64)
	SetFieldInfosFiles(fieldInfosFiles map[string]struct{})
	SetDocValuesUpdatesFiles(files map[int]map[string]struct{})
	Clone() SegmentCommitInfo
	GetId() []byte
	SizeInBytes() (int64, error)
	AdvanceDelGen()
	GetBufferedDeletesGen() int64
	GetFieldInfosFiles() map[string]struct{}
	GetDocValuesUpdatesFiles() map[int]map[string]struct{}
}

// segmentCommitInfo
// Embeds a [read-only] SegmentInfo and adds per-commit fields.
// lucene.experimental
type segmentCommitInfo struct {
	// The SegmentInfo that we wrap.
	info SegmentInfo

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

func NewSegmentCommitInfo(info SegmentInfo, delCount, softDelCount int,
	delGen, fieldInfosGen, docValuesGen int64, id []byte) SegmentCommitInfo {

	return newSegmentCommitInfo(info, delCount, softDelCount, delGen, fieldInfosGen, docValuesGen, id)
}

func newSegmentCommitInfo(info SegmentInfo, delCount, softDelCount int,
	delGen, fieldInfosGen, docValuesGen int64, id []byte) *segmentCommitInfo {

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

	return &segmentCommitInfo{
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

func (s *segmentCommitInfo) Info() SegmentInfo {
	return s.info
}

// HasDeletions Returns true if there are any deletions for the segment at this commit.
func (s *segmentCommitInfo) HasDeletions() bool {
	return s.delGen != -1
}

// HasFieldUpdates Returns true if there are any field updates for the segment in this commit.
func (s *segmentCommitInfo) HasFieldUpdates() bool {
	return s.fieldInfosGen != -1
}

// GetNextFieldInfosGen Returns the next available generation number of the FieldInfos files.
func (s *segmentCommitInfo) GetNextFieldInfosGen() int64 {
	return s.nextWriteFieldInfosGen
}

// GetFieldInfosGen Returns the generation number of the field infos file or -1 if there are no field updates yet.
func (s *segmentCommitInfo) GetFieldInfosGen() int64 {
	return s.fieldInfosGen
}

// GetNextDocValuesGen Returns the next available generation number of the DocValues files.
func (s *segmentCommitInfo) GetNextDocValuesGen() int64 {
	return s.nextWriteDocValuesGen
}

// GetDocValuesGen Returns the generation number of the DocValues file or -1 if there are no doc-values updates yet.
func (s *segmentCommitInfo) GetDocValuesGen() int64 {
	return s.docValuesGen
}

// GetNextDelGen Returns the next available generation number of the live docs file.
func (s *segmentCommitInfo) GetNextDelGen() int64 {
	return s.nextWriteDelGen
}

// GetDelGen Returns generation number of the live docs file or -1 if there are no deletes yet.
func (s *segmentCommitInfo) GetDelGen() int64 {
	return s.delGen
}

// GetDelCount Returns the number of deleted docs in the segment.
func (s *segmentCommitInfo) GetDelCount() int {
	return s.delCount
}

func (s *segmentCommitInfo) GetDelCountWithSoftDeletes(includeSoftDeletes bool) int {
	if includeSoftDeletes {
		return s.GetDelCount() + s.GetSoftDelCount()
	}
	return s.GetDelCount()
}

// GetSoftDelCount Returns the number of only soft-deleted docs.
func (s *segmentCommitInfo) GetSoftDelCount() int {
	return s.softDelCount
}

func (s *segmentCommitInfo) SetDelCount(delCount int) {
	s.delCount = delCount
}

func (s *segmentCommitInfo) SetSoftDelCount(softDelCount int) {
	s.softDelCount = softDelCount
}

func (s *segmentCommitInfo) Files() (map[string]struct{}, error) {
	files := s.info.Files()

	// Must separately add any live docs files:
	if _, err := s.info.GetCodec().LiveDocsFormat().Files(context.TODO(), s, files); err != nil {
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

func (s *segmentCommitInfo) GetNextWriteDelGen() int64 {
	return s.nextWriteDelGen
}

func (s *segmentCommitInfo) SetNextWriteDelGen(v int64) {
	s.nextWriteDelGen = v
}

func (s *segmentCommitInfo) GetNextWriteFieldInfosGen() int64 {
	return s.nextWriteFieldInfosGen
}

func (s *segmentCommitInfo) SetNextWriteFieldInfosGen(v int64) {
	s.nextWriteFieldInfosGen = v
}

func (s *segmentCommitInfo) GetNextWriteDocValuesGen() int64 {
	return s.nextWriteDocValuesGen
}

func (s *segmentCommitInfo) SetNextWriteDocValuesGen(v int64) {
	s.nextWriteDocValuesGen = v
}

func (s *segmentCommitInfo) SetFieldInfosFiles(fieldInfosFiles map[string]struct{}) {
	s.fieldInfosFiles = map[string]struct{}{}
	for file := range fieldInfosFiles {
		segmentName := s.info.NamedForThisSegment(file)
		s.fieldInfosFiles[segmentName] = struct{}{}
	}
}

func (s *segmentCommitInfo) SetDocValuesUpdatesFiles(files map[int]map[string]struct{}) {
	s.dvUpdatesFiles = map[int]map[string]struct{}{}
	for k, values := range files {
		s.dvUpdatesFiles[k] = maps.Clone(values)
	}
}

func (s *segmentCommitInfo) Clone() SegmentCommitInfo {
	other := newSegmentCommitInfo(s.info, s.delCount, s.softDelCount, s.delGen, s.fieldInfosGen, s.docValuesGen, s.GetId())
	// Not clear that we need to carry over nextWriteDelGen
	// (i.e. do we ever clone after a failed write and
	// before the next successful write?), but just do it to
	// be safe:

	other.nextWriteDelGen = s.nextWriteDelGen
	other.nextWriteFieldInfosGen = s.nextWriteFieldInfosGen
	other.nextWriteDocValuesGen = s.nextWriteDocValuesGen

	for k, files := range s.dvUpdatesFiles {
		other.dvUpdatesFiles[k] = maps.Clone(files)
	}

	for k := range s.fieldInfosFiles {
		other.fieldInfosFiles[k] = struct{}{}
	}

	return other
}

func (s *segmentCommitInfo) GetId() []byte {
	if len(s.id) > 0 {
		return slices.Clone(s.id)
	}
	return nil
}

func (s *segmentCommitInfo) SizeInBytes() (int64, error) {
	if s.sizeInBytes == -1 {
		sum := int64(0)

		files, _ := s.Files()
		for fileName := range files {
			fileLength, err := s.info.Dir().FileLength(nil, fileName)
			if err != nil {
				return 0, err
			}
			sum += fileLength
		}
		s.sizeInBytes = sum
	}

	return s.sizeInBytes, nil
}

// AdvanceDelGen
// Called when we succeed in writing deletes
func (s *segmentCommitInfo) AdvanceDelGen() {
	s.delGen = s.nextWriteDelGen
	s.nextWriteDelGen = s.delGen + 1
	s.generationAdvanced()
}

func (s *segmentCommitInfo) generationAdvanced() {
	s.sizeInBytes = -1
	r, _ := uuid.NewRandom()
	s.id = []byte(r.String())
}

func (s *segmentCommitInfo) GetBufferedDeletesGen() int64 {
	return s.bufferedDeletesGen
}

func (s *segmentCommitInfo) GetFieldInfosFiles() map[string]struct{} {
	return s.fieldInfosFiles
}

func (s *segmentCommitInfo) GetDocValuesUpdatesFiles() map[int]map[string]struct{} {
	return s.dvUpdatesFiles
}
