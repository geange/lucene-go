package codecs

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

type FieldInfosFormat interface {

	// Read the FieldInfos previously written with write.
	Read(directory store.Directory, segmentInfo *index.SegmentInfo,
		segmentSuffix string, ctx *store.IOContext) (*index.FieldInfos, error)

	// Write Writes the provided FieldInfos to the directory.
	Write(directory store.Directory, segmentInfo *index.SegmentInfo,
		segmentSuffix string, infos *index.FieldInfos, context *store.IOContext) error
}
