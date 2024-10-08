package index

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.DirectoryReader = &StandardDirectoryReader{}

// StandardDirectoryReader
// Default implementation of DirectoryReader.
type StandardDirectoryReader struct {
	*baseDirectoryReader

	writer          *IndexWriter
	segmentInfos    *SegmentInfos
	applyAllDeletes bool
	writeAllDeletes bool
}

// NewStandardDirectoryReader
// package private constructor, called only from static open() methods
func NewStandardDirectoryReader(directory store.Directory, readers []index.IndexReader, writer *IndexWriter,
	sis *SegmentInfos, compareFunc CompareLeafReader, applyAllDeletes, writeAllDeletes bool) (*StandardDirectoryReader, error) {

	reader, err := newBaseDirectoryReader(directory, readers, compareFunc)
	if err != nil {
		return nil, err
	}

	return &StandardDirectoryReader{
		baseDirectoryReader: reader,
		writer:              writer,
		segmentInfos:        sis,
		applyAllDeletes:     applyAllDeletes,
		writeAllDeletes:     writeAllDeletes,
	}, nil
}

type CompareIndexReader func(a, b index.IndexReader) int
type CompareLeafReader func(a, b index.LeafReader) int

// OpenDirectoryReader
// called from DirectoryReader.open(...) methods
func OpenDirectoryReader(ctx context.Context, directory store.Directory,
	commit IndexCommit, compareFunc CompareLeafReader) (index.DirectoryReader, error) {

	segmentsFile := NewFindSegmentsFile[index.DirectoryReader](directory)
	fnDoBody := func(ctx context.Context, segmentFileName string) (index.DirectoryReader, error) {
		sis, err := ReadCommit(ctx, segmentsFile.directory, segmentFileName)
		if err != nil {
			return nil, err
		}

		readers := make([]index.IndexReader, sis.Size())

		for i := sis.Size() - 1; i >= 0; i-- {
			reader, err := NewSegmentReader(ctx, sis.Info(i), sis.getIndexCreatedVersionMajor(), store.READ)
			if err != nil {
				return nil, err
			}
			readers[i] = reader
		}

		// This may throw CorruptIndexException if there are too many docs, so
		// it must be inside try clause so we close readers in that case:
		reader, err := NewStandardDirectoryReader(directory, readers, nil, sis, compareFunc,
			false, false)
		if err != nil {
			return nil, err
		}
		return reader, nil
	}
	segmentsFile.SetFuncDoBody(fnDoBody)

	reader, err := segmentsFile.RunWithCommit(ctx, commit)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// OpenStandardDirectoryReader
// Used by near real-time search
func OpenStandardDirectoryReader(writer *IndexWriter, readerFunction func(index.SegmentCommitInfo) (*SegmentReader, error),
	infos *SegmentInfos, applyAllDeletes, writeAllDeletes bool) (index.DirectoryReader, error) {

	// IndexWriter synchronizes externally before calling
	// us, which ensures infos will not change; so there's
	// no need to process segments in reverse order
	numSegments := infos.Size()

	readers := make([]index.IndexReader, 0)
	dir := writer.GetDirectory()
	segmentInfos := infos.Clone()
	infosUpto := 0

	for i := 0; i < numSegments; i++ {
		// NOTE: important that we use infos not
		// segmentInfos here, so that we are passing the
		// actual instance of SegmentInfoPerCommit in
		// IndexWriter's segmentInfos:
		info := infos.Info(i)
		//assert info.info.dir == dir;
		reader, err := readerFunction(info)
		if err != nil {
			return nil, err
		}
		if reader.NumDocs() > 0 || writer.GetConfig().mergePolicy.KeepFullyDeletedSegment(func() index.CodecReader {
			return reader
		}) {
			// Steal the ref:
			readers = append(readers, reader)
			infosUpto++
		} else {
			if err := reader.DecRef(); err != nil {
				return nil, err
			}
			segmentInfos.Remove(infosUpto)
		}
	}

	if err := writer.IncRefDeleter(segmentInfos); err != nil {
		return nil, err
	}

	sorter := writer.GetConfig().GetLeafSorter()

	result, err := NewStandardDirectoryReader(dir, readers, writer,
		segmentInfos, sorter, applyAllDeletes, writeAllDeletes)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *StandardDirectoryReader) GetVersion() int64 {
	return s.segmentInfos.GetVersion()
}

func (s *StandardDirectoryReader) IsCurrent(ctx context.Context) (bool, error) {
	if s.writer == nil || s.writer.IsClosed() {
		// Fully read the segments file: this ensures that it's
		// completely written so that if
		// IndexWriter.prepareCommit has been called (but not
		// yet commit), then the reader will still see itself as
		// current:
		sis, err := ReadLatestCommit(ctx, s.directory)
		if err != nil {
			return false, err
		}

		// we loaded SegmentInfos from the directory
		return sis.GetVersion() == s.segmentInfos.GetVersion(), nil
	}
	return s.writer.nrtIsCurrent(s.segmentInfos), nil
}

func (s *StandardDirectoryReader) GetIndexCommit() (index.IndexCommit, error) {
	return NewReaderCommit(s, s.segmentInfos, s.directory)
}

func (s *StandardDirectoryReader) GetSegmentInfos() *SegmentInfos {
	return s.segmentInfos
}

var _ index.IndexCommit = &ReaderCommit{}

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

func (r *ReaderCommit) CompareTo(commit index.IndexCommit) int {
	gen := r.GetGeneration()
	comgen := commit.GetGeneration()
	return Compare(gen, comgen)
}

func (r *ReaderCommit) GetReader() index.DirectoryReader {
	return r.reader
}
