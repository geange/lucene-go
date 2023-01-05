package index

import "github.com/geange/lucene-go/core/store"

// SegmentInfoFormat Expert: Controls the format of the
// SegmentInfo (segment metadata file).
// @see SegmentInfo
// @lucene.experimental
type SegmentInfoFormat interface {
	// Read {@link SegmentInfo} data from a directory.
	// @param directory directory to read from
	// @param segmentName name of the segment to read
	// @param segmentID expected identifier for the segment
	// @return infos instance to be populated with data
	// @throws IOException If an I/O error occurs
	Read(dir store.Directory, segmentName string, segmentID []byte, context *store.IOContext) (*SegmentInfo, error)

	// Write {@link SegmentInfo} data.
	// The codec must add its SegmentInfo filename(s) to {@code info} before doing i/o.
	// @throws IOException If an I/O error occurs
	Write(dir store.Directory, info *SegmentInfo, ioContext *store.IOContext) error
}
