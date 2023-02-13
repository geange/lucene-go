package index

import (
	"github.com/geange/lucene-go/core/util"
	"io"
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
//   - Version counts how often the index has been changed by adding or deleting documents.
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

	// Counts how often the index has been changed.
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
	infoStream io.Writer

	// Id for this commit; only written starting with Lucene 5.0
	id []byte

	// Which Lucene version wrote this commit.
	luceneVersion *util.Version

	// Version of the oldest segment in the index, or null if there are no segments.
	minSegmentLuceneVersion *util.Version

	// The Lucene version major that was used to create the index.
	indexCreatedVersionMajor int
}
