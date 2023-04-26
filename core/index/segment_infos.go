package index

import (
	"errors"
	"fmt"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"strconv"
	"strings"
)

const (
	// VERSION_70 The version that added information about the Lucene version at the time when the index has been created.
	VERSION_70 = 7

	// VERSION_72 The version that updated segment name counter to be long instead of int.
	VERSION_72 = 8

	// VERSION_74 The version that recorded softDelCount
	VERSION_74 = 9

	// VERSION_86 The version that recorded SegmentCommitInfo IDs
	VERSION_86      = 10
	VERSION_CURRENT = VERSION_86
)

// SegmentInfos A collection of segmentInfo objects with methods for operating on those segments in relation to the file system.
// The active segments in the index are stored in the segment info file, segments_N. There may be one or more segments_N files in the index; however, the one with the largest generation is the active one (when older segments_N files are present it's because they temporarily cannot be deleted, or a custom IndexDeletionPolicy is in use). This file lists each segment by name and has details about the codec and generation of deletes.
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
	luceneVersion *util.Version

	// Version of the oldest segment in the index, or null if there are no segments.
	minSegmentLuceneVersion *util.Version

	// The Lucene version major that was used to create the index.
	indexCreatedVersionMajor int
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

// Changed Call this before committing if changes have been made to the segments.
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

// Clone Returns a copy of this instance, also copying each SegmentInfo.
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

func (i *SegmentInfos) Replace(other *SegmentInfos) {
	i.rollbackSegmentInfos(other.AsList())
	i.lastGeneration = other.lastGeneration
}

func (i *SegmentInfos) AsList() []*SegmentCommitInfo {
	return i.segments
}

func (i *SegmentInfos) AddAll(sis []*SegmentCommitInfo) {
	for _, si := range sis {
		i.Add(si)
	}
}

func (i *SegmentInfos) rollbackSegmentInfos(list []*SegmentCommitInfo) {
	i.segments = i.segments[:0]
	i.AddAll(list)
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

func ReadCommit(directory store.Directory, segmentFileName string) (*SegmentInfos, error) {
	generation, err := GenerationFromSegmentsFileName(segmentFileName)
	if err != nil {
		return nil, err
	}

	input, err := store.OpenChecksumInput(directory, segmentFileName, nil)
	if err != nil {
		return nil, err
	}
	return ReadCommitFromChecksum(directory, input, generation)
}

// ReadCommitFromChecksum Read the commit from the provided ChecksumIndexInput.
func ReadCommitFromChecksum(directory store.Directory,
	input store.ChecksumIndexInput, generation int64) (*SegmentInfos, error) {

	format := -1

	// NOTE: as long as we want to throw indexformattooold (vs corruptindexexception), we need
	// to read the magic ourselves.
	magic, err := input.ReadUint32()
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
	ID_LENGTH := 16
	id := make([]byte, ID_LENGTH)
	_, err = input.Read(id)
	if err != nil {
		return nil, err
	}

	_, err = utils.CheckIndexHeaderSuffix(input, strconv.FormatInt(generation, 36))
	if err != nil {
		return nil, err
	}

	n1, err := input.ReadUvarint()
	if err != nil {
		return nil, err
	}
	n2, err := input.ReadUvarint()
	if err != nil {
		return nil, err
	}
	n3, err := input.ReadUvarint()
	if err != nil {
		return nil, err
	}
	luceneVersion := util.NewVersion(int(n1), int(n2), int(n3))
	indexCreatedVersion, err := input.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if luceneVersion.Major < int(indexCreatedVersion) {
		return nil, fmt.Errorf(
			"creation version [%d.x] can't be greater than the version that wrote the segment infos: [%d]",
			indexCreatedVersion, luceneVersion,
		)
	}

	if int(indexCreatedVersion) < util.VersionLast.Major-1 {
		return nil, errors.New("lucene only supports reading the current and previous major versions")
	}

	infos := NewSegmentInfos(int(indexCreatedVersion))
	infos.id = id
	infos.generation = generation
	infos.lastGeneration = generation
	infos.luceneVersion = luceneVersion

	version, err := input.ReadUint64()
	if err != nil {
		return nil, err
	}
	infos.version = int64(version)

	if format > VERSION_70 {
		count, err := input.ReadUvarint()
		if err != nil {
			return nil, err
		}
		infos.counter = int64(count)
	} else {
		count, err := input.ReadUint32()
		if err != nil {
			return nil, err
		}
		infos.counter = int64(count)
	}

	numSegments, err := input.ReadUint32()
	if err != nil {
		return nil, err
	}

	if numSegments > 0 {
		n1, err := input.ReadUvarint()
		if err != nil {
			return nil, err
		}
		n2, err := input.ReadUvarint()
		if err != nil {
			return nil, err
		}
		n3, err := input.ReadUvarint()
		if err != nil {
			return nil, err
		}
		infos.minSegmentLuceneVersion = util.NewVersion(int(n1), int(n2), int(n3))
	} else {
		// else leave as null: no segments
	}

	totalDocs := 0
	for seg := 0; seg < int(numSegments); seg++ {
		segName, err := input.ReadString()
		if err != nil {
			return nil, err
		}
		segmentID := make([]byte, ID_LENGTH)
		if _, err := input.Read(segmentID); err != nil {
			return nil, err
		}
		codec, err := ReadCodec(input)
		if err != nil {
			return nil, err
		}
		info, err := codec.SegmentInfoFormat().Read(directory, segName, segmentID, nil)
		if err != nil {
			return nil, err
		}
		info.SetCodec(codec)

		maxDoc, err := info.MaxDoc()
		if err != nil {
			return nil, err
		}
		totalDocs += maxDoc

		delGen, err := input.ReadUint64()
		if err != nil {
			return nil, err
		}
		delCount, err := input.ReadUint32()
		if err != nil {
			return nil, err
		}
		if delCount < 0 || int(delCount) > maxDoc {
			return nil, errors.New("invalid deletion count")
			//throw new CorruptIndexException("invalid deletion count: " + delCount + " vs maxDoc=" + info.maxDoc(), input);
		}
		fieldInfosGen, err := input.ReadUint64()
		if err != nil {
			return nil, err
		}
		dvGen, err := input.ReadUint64()
		if err != nil {
			return nil, err
		}
		softDelCount := 0
		if format > VERSION_72 {
			n, err := input.ReadUint32()
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
				_, err := input.Read(sciId)
				if err != nil {
					return nil, err
				}
				break
			case 0:
				sciId = nil
				break
			default:
				//throw new CorruptIndexException("invalid SegmentCommitInfo ID marker: " + marker, input);
			}
		} else {
			sciId = nil
		}
		siPerCommit := NewSegmentCommitInfo(info, int(delCount), softDelCount, int64(delGen), int64(fieldInfosGen), int64(dvGen), sciId)
		setOfStrings, err := input.ReadSetOfStrings()
		if err != nil {
			return nil, err
		}
		siPerCommit.SetFieldInfosFiles(setOfStrings)
		dvUpdateFiles := make(map[int]map[string]struct{})
		numDVFields, err := input.ReadUint32()
		if err != nil {
			return nil, err
		}
		if numDVFields == 0 {
			dvUpdateFiles = map[int]map[string]struct{}{}
		} else {
			values := make(map[int]map[string]struct{})
			for i := 0; i < int(numDVFields); i++ {
				num, err := input.ReadUint32()
				if err != nil {
					return nil, err
				}
				strs, err := input.ReadSetOfStrings()
				if err != nil {
					return nil, err
				}
				values[int(num)] = strs
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
			//throw new CorruptIndexException("segments file recorded minSegmentLuceneVersion=" + infos.minSegmentLuceneVersion + " but segment=" + info + " has older version=" + segmentVersion, input);
		}

		if infos.indexCreatedVersionMajor >= 7 && segmentVersion.Major < infos.indexCreatedVersionMajor {
			return nil, errors.New("version too new")
			//throw new CorruptIndexException("segments file recorded indexCreatedVersionMajor=" + infos.indexCreatedVersionMajor + " but segment=" + info + " has older version=" + segmentVersion, input);
		}

		if infos.indexCreatedVersionMajor >= 7 && info.GetMinVersion() == nil {
			return nil, fmt.Errorf(
				"segments infos must record minVersion with indexCreatedVersionMajor=%d",
				infos.indexCreatedVersionMajor)
			//throw new CorruptIndexException("segments infos must record minVersion with indexCreatedVersionMajor=" + infos.indexCreatedVersionMajor, input);
		}
	}

	return infos, nil
}

func ReadCodec(input store.DataInput) (Codec, error) {
	name, err := input.ReadString()
	if err != nil {
		return nil, err
	}
	codec := ForName(name)
	return codec, nil
}

// ReadLatestCommit Find the latest commit (segments_N file) and load all SegmentCommitInfos.
func ReadLatestCommit(directory store.Directory) (*SegmentInfos, error) {
	file := &FindSegmentsFile{
		directory: directory,
	}
	file.DoBody = func(segmentFileName string) (any, error) {
		return ReadCommit(file.directory, segmentFileName)
	}

	infos, err := file.Run()
	if err != nil {
		return nil, err
	}
	return infos.(*SegmentInfos), nil
}

// GenerationFromSegmentsFileName Parse the generation off the segments file name and return it.
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

// FindSegmentsFile Utility class for executing code that needs to do something with the current segments file.
// This is necessary with lock-less commits because from the time you locate the current segments file name,
// until you actually open it, read its contents, or check modified time, etc., it could have been deleted
// due to a writer commit finishing.
type FindSegmentsFile struct {
	directory store.Directory

	DoBody func(segmentFileName string) (any, error)
}

func NewFindSegmentsFile(directory store.Directory) *FindSegmentsFile {
	return &FindSegmentsFile{directory: directory}
}

func (f *FindSegmentsFile) Run() (any, error) {
	return f.RunV1(nil)
}

func (f *FindSegmentsFile) RunV1(commit IndexCommit) (any, error) {
	if commit != nil {
		if commit.GetDirectory() != f.directory {
			return nil, errors.New("the specified commit does not match the specified Directory")
		}
		return f.DoBody(commit.GetSegmentsFileName())
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
		files, err := f.directory.ListAll()
		if err != nil {
			return nil, err
		}

		gen := GetLastCommitGeneration(files)
		if gen == -1 {
			return nil, errors.New("no segments* file found in directory")
		} else if gen > lastGen {
			segmentFileName := FileNameFromGeneration(SEGMENTS, "", gen)

			value, err := f.DoBody(segmentFileName)
			if err != nil {
				return nil, err
			}
			return value, nil
		}
	}
}

// GetLastCommitGeneration Get the generation of the most recent commit to the list of index files
// (N in the segments_N file).
// Params: files – -- array of file names to check
func GetLastCommitGeneration(files []string) int64 {
	max := int64(-1)
	for _, file := range files {
		if strings.HasPrefix(file, SEGMENTS) && file != OLD_SEGMENTS_GEN {
			gen, _ := GenerationFromSegmentsFileName(file)
			if gen > max {
				max = gen
			}
		}
	}
	return max
}

func GetLastCommitSegmentsFileName(files []string) string {
	return FileNameFromGeneration(SEGMENTS, "", GetLastCommitGeneration(files))
}
