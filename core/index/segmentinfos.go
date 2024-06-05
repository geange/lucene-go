package index

import (
	"context"
	"errors"
	"fmt"
	codecUtil "github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/util"
	"strconv"
	"strings"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
)

const (
	// VERSION_70
	// The version that added information about the Lucene version at the time when the index has been created.
	VERSION_70 = 7

	// VERSION_72
	// The version that updated segment name counter to be long instead of int.
	VERSION_72 = 8

	// VERSION_74
	// The version that recorded softDelCount
	VERSION_74 = 9

	// VERSION_86
	// The version that recorded SegmentCommitInfo IDs
	VERSION_86      = 10
	VERSION_CURRENT = VERSION_86
)

// SegmentInfos
// A collection of segmentInfo objects with methods for operating on those segments in relation to the file system.
// The active segments in the index are stored in the segment info file, segments_N.
// There may be one or more segments_N files in the index; however, the one with the largest generation is
// the active one (when older segments_N files are present it's because they temporarily cannot be deleted,
// or a custom IndexDeletionPolicy is in use). This file lists each segment by name and has details about
// the codec and generation of deletes.
//
// Files:
//   - segments_N: Header, LuceneVersion, Version, NameCounter, SegCount, MinSegmentLuceneVersion, <SegName, SegID, SegCodec, DelGen, DeletionCount, FieldInfosGen, DocValuesGen, UpdatesFiles>SegCount, CommitUserData, Footer
//
// Data types:
//   - Header --> IndexHeader
//   - LuceneVersion --> Which Lucene code Version was used for this commit, written as three vInt: major, minor, bugfix
//   - MinSegmentLuceneVersion --> Lucene code Version of the oldest segment, written as three vInt: major, minor, bugfix; this is only written only if there's at least one segment
//   - NameCounter, SegCount, DeletionCount --> Int32
//   - Generation, Version, DelGen, Checksum, FieldInfosGen, DocValuesGen --> Int64
//   - SegID --> Int8ID_LENGTH
//   - SegName, SegCodec --> String
//   - CommitUserData --> Map<String,String>
//   - UpdatesFiles --> Map<Int32, Set<String>>
//   - Footer --> CodecFooter
//
// Field Descriptions:
//   - Version counts how often the index has been Changed by adding or deleting documents.
//   - NameCounter is used to generate names for new segment files.
//   - SegName is the name of the segment, and is used as the file name prefix for all of the files that compose the segment's index.
//   - DelGen is the generation count of the deletes file. If this is -1, there are no deletes. Anything above zero means there are deletes stored by LiveDocsFormat.
//   - DeletionCount records the number of deleted documents in this segment.
//   - SegCodec is the name of the Codec that encoded this segment.
//   - SegID is the identifier of the Codec that encoded this segment.
//   - CommitUserData stores an optional user-supplied opaque Map<String,String> that was passed to IndexWriter.setLiveCommitData(Iterable).
//   - FieldInfosGen is the generation count of the fieldInfos file. If this is -1, there are no updates to the fieldInfos in that segment. Anything above zero means there are updates to fieldInfos stored by FieldInfosFormat .
//   - DocValuesGen is the generation count of the updatable DocValues. If this is -1, there are no updates to DocValues in that segment. Anything above zero means there are updates to DocValues stored by DocValuesFormat.
//   - UpdatesFiles stores the set of files that were updated in that segment per field.
//
// lucene.experimental
type SegmentInfos struct {

	// Used to name new segments.
	counter int64

	// Counts how often the index has been Changed.
	version int64

	// generation of the "segments_N" for the next commit
	generation int64

	// generation of the "segments_N" file we last successfully read
	// or wrote; this is normally the same as generation except if
	// there was an IOException that had interrupted a commit
	lastGeneration int64

	// Opaque Map<String, String> that user can specify during IndexWriter.commit
	userData map[string]string

	segments []*SegmentCommitInfo

	// If non-null, information about loading segments_N files will be printed here. @see #setInfoStream.
	//infoStream io.Writer

	// Id for this commit; only written starting with Lucene 5.0
	id []byte

	// Which Lucene version wrote this commit.
	luceneVersion *version.Version

	// Version of the oldest segment in the index, or null if there are no segments.
	minSegmentLuceneVersion *version.Version

	// The Lucene version major that was used to create the index.
	indexCreatedVersionMajor int

	// Only true after prepareCommit has been called and
	// before finishCommit is called
	pendingCommit bool
}

func NewSegmentInfos(indexCreatedVersionMajor int) *SegmentInfos {
	return &SegmentInfos{
		counter:        0,
		version:        0,
		generation:     0,
		lastGeneration: 0,
		userData:       map[string]string{},
		segments:       make([]*SegmentCommitInfo, 0),
		//infoStream:               nil,
		id:                       nil,
		luceneVersion:            nil,
		minSegmentLuceneVersion:  nil,
		indexCreatedVersionMajor: indexCreatedVersionMajor,
	}
}

func (i *SegmentInfos) getIndexCreatedVersionMajor() int {
	return i.indexCreatedVersionMajor
}

// Changed
// Call this before committing if changes have been made to the segments.
func (i *SegmentInfos) Changed() {
	i.version++
}

func (i *SegmentInfos) GetSegmentsFileName() string {
	return FileNameFromGeneration(SEGMENTS, "", i.lastGeneration)
}

func (i *SegmentInfos) GetUserData() map[string]string {
	return i.userData
}

func (i *SegmentInfos) GetGeneration() int64 {
	return i.generation
}

func (i *SegmentInfos) Size() int {
	return len(i.segments)
}

func (i *SegmentInfos) prepareCommit(ctx context.Context, dir store.Directory) error {
	if i.pendingCommit {
		return errors.New("prepareCommit was already called")
	}
	return i.writeDir(ctx, dir)
}

func (i *SegmentInfos) Files(includeSegmentsFile bool) (map[string]struct{}, error) {
	files := make(map[string]struct{})
	if includeSegmentsFile {
		segmentFileName := i.GetSegmentsFileName()
		if segmentFileName != "" {
			files[segmentFileName] = struct{}{}
		}
	}

	for _, info := range i.segments {
		infoFiles, err := info.Files()
		if err != nil {
			return nil, err
		}
		for file := range infoFiles {
			files[file] = struct{}{}
		}
	}

	return files, nil
}

func (i *SegmentInfos) Info(j int) *SegmentCommitInfo {
	return i.segments[j]
}

func (i *SegmentInfos) SetNextWriteGeneration(generation int64) {
	i.generation = generation
}

func (i *SegmentInfos) Add(si *SegmentCommitInfo) error {
	if i.indexCreatedVersionMajor >= 7 && si.info.minVersion == nil {
		return errors.New("all segments must record the minVersion for indices created on or after Lucene 7")
	}
	i.segments = append(i.segments, si)
	return nil
}

func (i *SegmentInfos) UpdateGenerationVersionAndCounter(other *SegmentInfos) {
	i.UpdateGeneration(other)
	i.version = other.version
	i.counter = other.counter
}

func (i *SegmentInfos) UpdateGeneration(other *SegmentInfos) {
	i.lastGeneration = other.lastGeneration
	i.generation = other.generation
}

func (i *SegmentInfos) CreateBackupSegmentInfos() []*SegmentCommitInfo {
	list := make([]*SegmentCommitInfo, 0, i.Size())
	for _, segment := range i.segments {
		list = append(list, segment.Clone())
	}
	return list
}

func (i *SegmentInfos) GetLastGeneration() int64 {
	return i.lastGeneration
}

// Clone
// Returns a copy of this instance, also copying each SegmentInfo.
func (i *SegmentInfos) Clone() *SegmentInfos {
	infos := &SegmentInfos{
		counter:                  i.counter,
		version:                  i.version,
		generation:               i.generation,
		lastGeneration:           i.lastGeneration,
		userData:                 map[string]string{},
		segments:                 []*SegmentCommitInfo{},
		id:                       make([]byte, len(i.id)),
		luceneVersion:            i.luceneVersion.Clone(),
		minSegmentLuceneVersion:  i.luceneVersion.Clone(),
		indexCreatedVersionMajor: i.indexCreatedVersionMajor,
	}

	for k, v := range i.userData {
		infos.userData[k] = v
	}

	for _, segment := range i.segments {
		infos.segments = append(infos.segments, segment.Clone())
	}
	return infos
}

func (s *SegmentInfos) Commit(ctx context.Context, dir store.Directory) error {
	if err := s.prepareCommit(ctx, dir); err != nil {
		return err
	}
	return s.finishCommit(ctx, dir)
}

func (s *SegmentInfos) finishCommit(ctx context.Context, dir store.Directory) error {
	src := FileNameFromGeneration(PENDING_SEGMENTS, "", s.generation)
	dest := FileNameFromGeneration(SEGMENTS, "", s.generation)
	return dir.Rename(ctx, src, dest)
}

func (s *SegmentInfos) writeIndexOutput(ctx context.Context, out store.IndexOutput) error {
	if err := codecUtil.WriteIndexHeader(ctx, out, "segments", VERSION_CURRENT,
		util.RandomId(), fmt.Sprintf("%d", s.generation)); err != nil {
		return err
	}

	if err := out.WriteUvarint(ctx, uint64(version.Last.Major())); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(version.Last.Minor())); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(version.Last.Bugfix())); err != nil {
		return err
	}
	//System.out.println(Thread.currentThread().getName() + ": now write " + out.getName() + " with version=" + version);

	if err := out.WriteUvarint(ctx, uint64(s.indexCreatedVersionMajor)); err != nil {
		return err
	}

	if err := out.WriteUint64(ctx, uint64(s.version)); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(s.counter)); err != nil {
		return err
	} // write counter
	if err := out.WriteUint32(ctx, uint32(s.Size())); err != nil {
		return err
	}

	if s.Size() > 0 {
		minSegmentVersion := &version.Version{}

		// We do a separate loop up front so we can write the minSegmentVersion before
		// any SegmentInfo; this makes it cleaner to throw IndexFormatTooOldExc at read time:
		for _, siPerCommit := range s.segments {
			segmentVersion := siPerCommit.info.GetVersion()
			if minSegmentVersion == nil || segmentVersion.OnOrAfter(minSegmentVersion) == false {
				minSegmentVersion = segmentVersion
			}
		}

		if err := out.WriteUvarint(ctx, uint64(minSegmentVersion.Major())); err != nil {
			return err
		}
		if err := out.WriteUvarint(ctx, uint64(minSegmentVersion.Minor())); err != nil {
			return err
		}
		if err := out.WriteUvarint(ctx, uint64(minSegmentVersion.Bugfix())); err != nil {
			return err
		}
	}

	// write infos
	for _, siPerCommit := range s.segments {
		si := siPerCommit.info
		if s.indexCreatedVersionMajor >= 7 && si.minVersion == nil {
			return fmt.Errorf("segments must record minVersion if they have been created on or after Lucene 7: %+v", si)
		}
		if err := out.WriteString(ctx, si.name); err != nil {
			return err
		}
		segmentID := si.GetID()
		if len(segmentID) != ID_LENGTH {
			return fmt.Errorf("cannot write segment: invalid id segment=%s id=%s", si.name, util.StringRandomId(segmentID))
		}
		if _, err := out.Write(segmentID); err != nil {
			return err
		}
		if err := out.WriteString(ctx, si.GetCodec().GetName()); err != nil {
			return err
		}
		if err := out.WriteUint64(ctx, uint64(siPerCommit.GetDelGen())); err != nil {
			return err
		}
		delCount := siPerCommit.GetDelCount()

		maxDoc, err := si.MaxDoc()
		if err != nil {
			return err
		}
		if delCount < 0 || delCount > maxDoc {
			return fmt.Errorf("cannot write segment: invalid maxDoc segment=%s maxDox=%d", si.name, delCount)
		}
		if err := out.WriteUint32(ctx, uint32(delCount)); err != nil {
			return err
		}
		if err := out.WriteUint64(ctx, uint64(siPerCommit.GetFieldInfosGen())); err != nil {
			return err
		}
		if err := out.WriteUint64(ctx, uint64(siPerCommit.GetDocValuesGen())); err != nil {
			return err
		}
		softDelCount := siPerCommit.GetSoftDelCount()
		if softDelCount < 0 || softDelCount > maxDoc {
			return fmt.Errorf("cannot write segment: invalid maxDoc segment=%s maxDoc=%d", si.name, softDelCount)
		}
		if err := out.WriteUint32(ctx, uint32(softDelCount)); err != nil {
			return err
		}
		// we ensure that there is a valid ID for this SCI just in case
		// this is manually upgraded outside of IW
		sciId := siPerCommit.GetId()
		if sciId != nil {
			if err := out.WriteByte(1); err != nil {
				return err
			}
			if _, err := out.Write(sciId); err != nil {
				return err
			}
		} else {
			if err := out.WriteByte(0); err != nil {
				return err
			}
		}

		if err := out.WriteSetOfStrings(ctx, siPerCommit.GetFieldInfosFiles()); err != nil {
			return err
		}
		dvUpdatesFiles := siPerCommit.GetDocValuesUpdatesFiles()
		if err := out.WriteUint32(ctx, uint32(len(dvUpdatesFiles))); err != nil {
			return err
		}

		for key, sets := range dvUpdatesFiles {
			if err := out.WriteUint32(ctx, uint32(key)); err != nil {
				return err
			}
			if err := out.WriteSetOfStrings(ctx, sets); err != nil {
				return err
			}
		}
	}

	if err := out.WriteMapOfStrings(ctx, s.userData); err != nil {
		return err
	}
	return codecUtil.WriteFooter(out)
}

func (s *SegmentInfos) writeDir(ctx context.Context, directory store.Directory) error {
	nextGeneration := s.getNextPendingGeneration()
	segmentFileName := FileNameFromGeneration(PENDING_SEGMENTS, "", nextGeneration)

	// Always advance the generation on write:
	s.generation = nextGeneration

	var segNOutput store.IndexOutput

	segNOutput, err := directory.CreateOutput(ctx, segmentFileName)
	if err != nil {
		return err
	}
	if err := s.writeIndexOutput(ctx, segNOutput); err != nil {
		return err
	}
	if err := segNOutput.Close(); err != nil {
		return err
	}
	return nil
}

func (i *SegmentInfos) Replace(other *SegmentInfos) error {
	if err := i.rollbackSegmentInfos(other.AsList()); err != nil {
		return err
	}
	i.lastGeneration = other.lastGeneration
	return nil
}

func (i *SegmentInfos) AsList() []*SegmentCommitInfo {
	return i.segments
}

func (i *SegmentInfos) AddAll(sis []*SegmentCommitInfo) error {
	for _, si := range sis {
		if err := i.Add(si); err != nil {
			return err
		}
	}
	return nil
}

func (i *SegmentInfos) rollbackSegmentInfos(list []*SegmentCommitInfo) error {
	i.segments = i.segments[:0]
	return i.AddAll(list)
}

func (i *SegmentInfos) TotalMaxDoc() int64 {
	count := 0
	for _, info := range i.segments {
		maxDoc, _ := info.info.MaxDoc()
		count += maxDoc
	}
	return int64(count)
}

func (i *SegmentInfos) GetVersion() int64 {
	return i.version
}

func (i *SegmentInfos) Remove(index int) {
	i.segments[index] = nil
}

// return generation of the next pending_segments_N that will be written
func (i *SegmentInfos) getNextPendingGeneration() int64 {
	if i.generation == -1 {
		return 1
	} else {
		return i.generation + 1
	}
}

func ReadCommit(ctx context.Context, directory store.Directory, segmentFileName string) (*SegmentInfos, error) {
	generation, err := GenerationFromSegmentsFileName(segmentFileName)
	if err != nil {
		return nil, err
	}

	input, err := store.OpenChecksumInput(directory, segmentFileName)
	if err != nil {
		return nil, err
	}
	return ReadCommitFromChecksumIndexInput(ctx, directory, input, generation)
}

const (
	ID_LENGTH = 16
)

// ReadCommitFromChecksumIndexInput
// Read the commit from the provided ChecksumIndexInput.
func ReadCommitFromChecksumIndexInput(ctx context.Context, directory store.Directory, input store.ChecksumIndexInput, generation int64) (*SegmentInfos, error) {

	format := -1

	// NOTE: as long as we want to throw indexformattooold (vs corruptindexexception), we need
	// to read the magic ourselves.
	magic, err := input.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	if int(magic) != utils.CODEC_MAGIC {
		//fmt.Println(magic)
		return nil, errors.New("indexFormat Too Old Exception")
	}

	format, err = utils.CheckHeaderNoMagic(input, "segments", VERSION_70, VERSION_CURRENT)
	if err != nil {
		return nil, err
	}

	id := make([]byte, ID_LENGTH)
	if _, err = input.Read(id); err != nil {
		return nil, err
	}

	if _, err = utils.CheckIndexHeaderSuffix(input, strconv.FormatInt(generation, 36)); err != nil {
		return nil, err
	}

	major, err := input.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	minor, err := input.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	bugfix, err := input.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	luceneVersion, err := version.New(
		version.WithMajor(uint8(major)),
		version.WithMinor(uint8(minor)),
		version.WithBugfix(uint8(bugfix)),
	)
	if err != nil {
		return nil, err
	}
	indexCreatedVersion, err := input.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	if uint64(luceneVersion.Major()) < (indexCreatedVersion) {
		formatStr := "creation version [%d.x] can't be greater than the version that wrote the segment infos: [%d]"
		return nil, fmt.Errorf(formatStr, indexCreatedVersion, luceneVersion)
	}

	if int(indexCreatedVersion) < int(version.Last.Major()-1) {
		return nil, errors.New("lucene only supports reading the current and previous major versions")
	}

	infos := NewSegmentInfos(int(indexCreatedVersion))
	infos.id = id
	infos.generation = generation
	infos.lastGeneration = generation
	infos.luceneVersion = luceneVersion

	ver, err := input.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}
	infos.version = int64(ver)

	if format > VERSION_70 {
		count, err := input.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		infos.counter = int64(count)
	} else {
		count, err := input.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		infos.counter = int64(count)
	}

	numSegments, err := input.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}

	if numSegments > 0 {
		// major, minor, bugfix
		major, err := input.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		minor, err := input.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		bugfix, err := input.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		minSegmentLuceneVersion, err := version.New(
			version.WithMajor(uint8(major)),
			version.WithMinor(uint8(minor)),
			version.WithBugfix(uint8(bugfix)),
		)
		if err != nil {
			return nil, err
		}
		infos.minSegmentLuceneVersion = minSegmentLuceneVersion
	} else {
		// else leave as null: no segments
	}

	totalDocs := 0
	for seg := 0; seg < int(numSegments); seg++ {
		segName, err := input.ReadString(ctx)
		if err != nil {
			return nil, err
		}
		segmentID := make([]byte, ID_LENGTH)
		if _, err := input.Read(segmentID); err != nil {
			return nil, err
		}
		codec, err := ReadCodec(ctx, input)
		if err != nil {
			return nil, err
		}
		info, err := codec.SegmentInfoFormat().Read(ctx, directory, segName, segmentID, nil)
		if err != nil {
			return nil, err
		}
		info.SetCodec(codec)

		maxDoc, err := info.MaxDoc()
		if err != nil {
			return nil, err
		}
		totalDocs += maxDoc

		delGen, err := input.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		delCount, err := input.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		if delCount < 0 || int(delCount) > maxDoc {
			return nil, errors.New("invalid deletion count")
			//throw new CorruptIndexException("invalid deletion count: " + delCount + " vs maxDoc=" + info.maxDoc(), input);
		}
		fieldInfosGen, err := input.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		dvGen, err := input.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		softDelCount := 0
		if format > VERSION_72 {
			n, err := input.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			softDelCount = int(n)
		}
		if softDelCount < 0 || softDelCount > maxDoc {
			return nil, errors.New("invalid deletion count")
			//throw new CorruptIndexException("invalid deletion count: " + softDelCount + " vs maxDoc=" + info.maxDoc(), input);
		}
		if softDelCount+int(delCount) > maxDoc {
			return nil, errors.New("invalid deletion count")
			//throw new CorruptIndexException("invalid deletion count: " + (softDelCount + delCount) + " vs maxDoc=" + info.maxDoc(), input);
		}
		var sciId []byte
		if format > VERSION_74 {
			marker, err := input.ReadByte()
			if err != nil {
				return nil, err
			}
			switch marker {
			case 1:
				sciId = make([]byte, ID_LENGTH)
				if _, err := input.Read(sciId); err != nil {
					return nil, err
				}
				break
			case 0:
				sciId = nil
				break
			default:
				return nil, fmt.Errorf("invalid SegmentCommitInfo ID marker: %d", marker)
			}
		} else {
			sciId = nil
		}
		siPerCommit := NewSegmentCommitInfo(info, int(delCount), softDelCount, int64(delGen), int64(fieldInfosGen), int64(dvGen), sciId)
		fieldInfosFiles, err := input.ReadSetOfStrings(ctx)
		if err != nil {
			return nil, err
		}
		siPerCommit.SetFieldInfosFiles(fieldInfosFiles)
		dvUpdateFiles := make(map[int]map[string]struct{})
		numDVFields, err := input.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		if numDVFields == 0 {
			dvUpdateFiles = map[int]map[string]struct{}{}
		} else {
			values := make(map[int]map[string]struct{})
			for i := 0; i < int(numDVFields); i++ {
				num, err := input.ReadUint32(ctx)
				if err != nil {
					return nil, err
				}
				strValues, err := input.ReadSetOfStrings(ctx)
				if err != nil {
					return nil, err
				}
				values[int(num)] = strValues
			}

			dvUpdateFiles = values
		}
		siPerCommit.SetDocValuesUpdatesFiles(dvUpdateFiles)
		if err := infos.Add(siPerCommit); err != nil {
			return nil, err
		}

		segmentVersion := info.GetVersion()

		if segmentVersion.OnOrAfter(infos.minSegmentLuceneVersion) == false {
			return nil, errors.New("version too old")
		}

		if infos.indexCreatedVersionMajor >= 7 && int(segmentVersion.Major()) < infos.indexCreatedVersionMajor {
			return nil, errors.New("version too new")
		}

		if infos.indexCreatedVersionMajor >= 7 && info.GetMinVersion() == nil {
			return nil, fmt.Errorf("segments infos must record minVersion with indexCreatedVersionMajor=%d",
				infos.indexCreatedVersionMajor)
		}
	}

	return infos, nil
}

func ReadCodec(ctx context.Context, input store.DataInput) (Codec, error) {
	name, err := input.ReadString(ctx)
	if err != nil {
		return nil, err
	}
	codec, exist := GetCodecByName(name)
	if !exist {
		return nil, fmt.Errorf("codec:%s not found", name)
	}
	return codec, nil
}

// ReadLatestCommit
// Find the latest commit (segments_N file) and load all SegmentCommitInfos.
func ReadLatestCommit(ctx context.Context, directory store.Directory) (*SegmentInfos, error) {
	file := &FindSegmentsFile[*SegmentInfos]{
		directory: directory,
	}

	file.fnDoBody = func(ctx context.Context, segmentFileName string) (*SegmentInfos, error) {
		return ReadCommit(ctx, file.directory, segmentFileName)
	}

	infos, err := file.Run(ctx)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// GenerationFromSegmentsFileName
// Parse the generation off the segments file name and return it.
func GenerationFromSegmentsFileName(fileName string) (int64, error) {
	if fileName == SEGMENTS {
		return 0, nil
	}

	if strings.HasPrefix(fileName, SEGMENTS) {
		v := fileName[len(SEGMENTS)+1:]
		return strconv.ParseInt(v, 10, 64)
	}

	return 0, fmt.Errorf("fileName '%s' is not a segments file", fileName)
}

// FindSegmentsFile
// Utility class for executing code that needs to do something with the current segments file.
// This is necessary with lock-less commits because from the time you locate the current segments file name,
// until you actually open it, read its contents, or check modified time, etc., it could have been deleted
// due to a writer commit finishing.
type FindSegmentsFile[T any] struct {
	directory store.Directory
	fnDoBody  func(ctx context.Context, segmentFileName string) (T, error)
}

func NewFindSegmentsFile[T any](directory store.Directory) *FindSegmentsFile[T] {
	return &FindSegmentsFile[T]{
		directory: directory,
	}
}

// Run
// Locate the most recent segments file and run doBody on it.
func (f *FindSegmentsFile[T]) Run(ctx context.Context) (T, error) {
	return f.RunWithCommit(ctx, nil)
}

// RunWithCommit
// Run doBody on the provided commit.
func (f *FindSegmentsFile[T]) RunWithCommit(ctx context.Context, commit IndexCommit) (T, error) {
	var _t T

	if commit != nil {
		if commit.GetDirectory() != f.directory {
			return _t, errors.New("the specified commit does not match the specified Directory")
		}
		return f.fnDoBody(ctx, commit.GetSegmentsFileName())
	}

	lastGen := int64(-1)
	gen := int64(-1)

	// Loop until we succeed in calling DoBody() without
	// hitting an IOException.  An IOException most likely
	// means an IW deleted our commit while opening
	// the time it took us to load the now-old infos files
	// (and segments files).  It's also possible it's a
	// true error (corrupt index).  To distinguish these,
	// on each retry we must see "forward progress" on
	// which generation we are trying to load.  If we
	// don't, then the original error is real and we throw
	// it.

	for {
		lastGen = gen
		files, err := f.directory.ListAll(ctx)
		if err != nil {
			return _t, err
		}

		gen, err = GetLastCommitGeneration(files)
		if err != nil {
			return _t, err
		}

		if gen == -1 {
			return _t, errors.New("no segments* file found in directory")
		} else if gen > lastGen {
			segmentFileName := FileNameFromGeneration(SEGMENTS, "", gen)

			value, err := f.fnDoBody(ctx, segmentFileName)
			if err != nil {
				return _t, err
			}
			return value, nil
		}
	}
}

func (f *FindSegmentsFile[T]) SetFuncDoBody(fnDoBody func(ctx context.Context, segmentFileName string) (T, error)) {
	f.fnDoBody = fnDoBody
}

// GetLastCommitGeneration
// Get the generation of the most recent commit to the list of index files
// (N in the segments_N file).
// Params: files – -- array of file names to check
func GetLastCommitGeneration(files []string) (int64, error) {
	maxGen := int64(-1)
	for _, file := range files {
		if strings.HasPrefix(file, SEGMENTS) && file != OLD_SEGMENTS_GEN {
			gen, err := GenerationFromSegmentsFileName(file)
			if err != nil {
				return 0, err
			}
			if gen > maxGen {
				maxGen = gen
			}
		}
	}
	return maxGen, nil
}

func GetLastCommitSegmentsFileName(files []string) (string, error) {
	generation, err := GetLastCommitGeneration(files)
	if err != nil {
		return "", err
	}
	return FileNameFromGeneration(SEGMENTS, "", generation), nil
}
