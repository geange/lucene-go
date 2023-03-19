package index

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
)

var _ DirectoryReader = &StandardDirectoryReader{}

// StandardDirectoryReader Default implementation of DirectoryReader.
type StandardDirectoryReader struct {
	*DirectoryReaderDefault

	writer          *IndexWriter
	segmentInfos    *SegmentInfos
	applyAllDeletes bool
	writeAllDeletes bool
}

// NewStandardDirectoryReader package private constructor, called only from static open() methods
func NewStandardDirectoryReader(directory store.Directory, readers []IndexReader, writer *IndexWriter,
	sis *SegmentInfos, leafSorter func(a, b IndexReader) int,
	applyAllDeletes, writeAllDeletes bool) (*StandardDirectoryReader, error) {

	reader, err := NewDirectoryReader(directory, readers, leafSorter)
	if err != nil {
		return nil, err
	}

	return &StandardDirectoryReader{
		DirectoryReaderDefault: reader,
		writer:                 writer,
		segmentInfos:           sis,
		applyAllDeletes:        applyAllDeletes,
		writeAllDeletes:        writeAllDeletes,
	}, nil
}

// OpenDirectoryReader
// called from DirectoryReader.open(...) methods
func OpenDirectoryReader(directory store.Directory,
	commit IndexCommit, leafSorter func(a, b IndexReader) int) (DirectoryReader, error) {

	reader, err := NewFindSegmentsFile(directory).RunV1(commit)
	if err != nil {
		return nil, err
	}
	return reader.(DirectoryReader), nil
}

// Used by near real-time search
func OpenDirectoryReaderV1(writer *IndexWriter,
	readerFunction func(*SegmentCommitInfo) *SegmentReader, infos *SegmentInfos,
	applyAllDeletes, writeAllDeletes bool) (*StandardDirectoryReader, error) {

	panic("")
}

func (s *StandardDirectoryReader) GetVersion() int64 {
	//ensureOpen();
	return s.segmentInfos.GetVersion()
}

func (s *StandardDirectoryReader) IsCurrent() (bool, error) {
	//ensureOpen();
	if s.writer == nil || s.writer.IsClosed() {
		// Fully read the segments file: this ensures that it's
		// completely written so that if
		// IndexWriter.prepareCommit has been called (but not
		// yet commit), then the reader will still see itself as
		// current:
		sis, err := ReadLatestCommit(s.directory)
		if err != nil {
			return false, err
		}

		// we loaded SegmentInfos from the directory
		return sis.GetVersion() == s.segmentInfos.GetVersion(), nil
	}
	return s.writer.nrtIsCurrent(s.segmentInfos), nil
}

func (s *StandardDirectoryReader) GetIndexCommit() (IndexCommit, error) {
	return NewReaderCommit(s, s.segmentInfos, s.directory)
}

var _ IndexCommit = &ReaderCommit{}

type ReaderCommit struct {
	segmentsFileName string
	files            map[string]struct{}
	dir              store.Directory
	generation       int64
	userData         map[string]string
	segmentCount     int
	reader           *StandardDirectoryReader
}

func NewReaderCommit(reader *StandardDirectoryReader,
	infos *SegmentInfos, dir store.Directory) (*ReaderCommit, error) {

	files, err := infos.Files(true)
	if err != nil {
		return nil, err
	}

	return &ReaderCommit{
		segmentsFileName: infos.GetSegmentsFileName(),
		files:            files,
		dir:              dir,
		generation:       infos.GetGeneration(),
		userData:         infos.GetUserData(),
		segmentCount:     infos.Size(),
		reader:           reader,
	}, nil
}

func (r *ReaderCommit) GetSegmentsFileName() string {
	return r.segmentsFileName
}

func (r *ReaderCommit) GetFileNames() (map[string]struct{}, error) {
	return r.files, nil
}

func (r *ReaderCommit) GetDirectory() store.Directory {
	return r.dir
}

func (r *ReaderCommit) Delete() error {
	return errors.New("this IndexCommit does not support deletions")
}

func (r *ReaderCommit) IsDeleted() bool {
	return false
}

func (r *ReaderCommit) GetSegmentCount() int {
	return r.segmentCount
}

func (r *ReaderCommit) GetGeneration() int64 {
	return r.generation
}

func (r *ReaderCommit) GetUserData() (map[string]string, error) {
	return r.userData, nil
}

func (r *ReaderCommit) CompareTo(commit IndexCommit) int {
	gen := r.GetGeneration()
	comgen := commit.GetGeneration()
	return Compare(gen, comgen)
}

func (r *ReaderCommit) GetReader() *StandardDirectoryReader {
	return r.reader
}
