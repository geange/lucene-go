package index

import (
	"github.com/geange/lucene-go/core/store"
)

type FieldInfosFormat interface {

	// Read the FieldInfos previously written with write.
	Read(directory store.Directory, segmentInfo *SegmentInfo,
		segmentSuffix string, ctx *store.IOContext) (*FieldInfos, error)

	// Write Writes the provided FieldInfos to the directory.
	Write(directory store.Directory, segmentInfo *SegmentInfo,
		segmentSuffix string, infos *FieldInfos, context *store.IOContext) error
}
