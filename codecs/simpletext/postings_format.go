package simpletext

import "github.com/geange/lucene-go/core/store"

const (
	POSTINGS_EXTENSION = "pst"
)

func getPostingsFileName(segment, segmentSuffix string) string {
	return store.SegmentFileName(segment, segmentSuffix, POSTINGS_EXTENSION)
}
