package index

import (
	"context"
	"strings"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type baseDirectoryReader struct {
	*baseCompositeReader

	directory store.Directory
}

func newBaseDirectoryReader(directory store.Directory,
	segmentReaders []index.IndexReader, leafSorter CompareLeafReader) (*baseDirectoryReader, error) {

	reader, err := newBaseCompositeReader(segmentReaders, leafSorter)
	if err != nil {
		return nil, err
	}

	return &baseDirectoryReader{
		baseCompositeReader: reader,
		directory:           directory,
	}, nil
}

func DirectoryReaderOpen(ctx context.Context, writer *IndexWriter) (index.DirectoryReader, error) {
	return DirectoryReaderOpenV1(ctx, writer, true, false)
}

func DirectoryReaderOpenV1(ctx context.Context, writer *IndexWriter, applyAllDeletes, writeAllDeletes bool) (index.DirectoryReader, error) {
	return writer.GetReader(ctx, applyAllDeletes, writeAllDeletes)
}

func (d *baseDirectoryReader) Directory() store.Directory {
	return d.directory
}

// IsIndexExists
// Returns true if an index likely exists at the specified directory.
// Note that if a corrupt index exists, or if an index in the process of committing
// Params: directory – the directory to check for an index
// Returns: true if an index exists; false otherwise
func IsIndexExists(dir store.Directory) (bool, error) {
	// LUCENE-2812, LUCENE-2727, LUCENE-4738: this logic will
	// return true in cases that should arguably be false,
	// such as only IW.prepareCommit has been called, or a
	// corrupt first commit, but it's too deadly to make
	// this logic "smarter" and risk accidentally returning
	// false due to various cases like file description
	// exhaustion, access denied, etc., because in that
	// case IndexWriter may delete the entire index.  It's
	// safer to err towards "index exists" than try to be
	// smart about detecting not-yet-fully-committed or
	// corrupt indices.  This means that IndexWriter will
	// throw an exception on such indices and the app must
	// resolve the situation manually:
	files, err := dir.ListAll(nil)
	if err != nil {
		return false, err
	}

	prefix := SEGMENTS + "_"
	for _, file := range files {
		if strings.HasPrefix(file, prefix) {
			return true, nil
		}
	}
	return false, nil
}

type DirectoryReaderBuilder struct {
}
