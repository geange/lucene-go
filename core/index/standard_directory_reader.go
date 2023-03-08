package index

import (
	"github.com/geange/lucene-go/core/store"
)

// StandardDirectoryReader Default implementation of DirectoryReader.
type StandardDirectoryReader struct {
	*DirectoryReader

	writer          *IndexWriter
	segmentInfos    *SegmentInfos
	applyAllDeletes bool
	writeAllDeletes bool
}

func (r *StandardDirectoryReader) Directory() store.Directory {
	return r.directory
}
