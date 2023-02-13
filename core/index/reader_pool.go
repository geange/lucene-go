package index

// ReaderPool Holds shared SegmentReader instances. IndexWriter uses SegmentReaders for 1) applying
// deletes/DV updates, 2) doing merges, 3) handing out a real-time reader. This pool reuses instances
// of the SegmentReaders in all these places if it is in "near real-time mode" (getReader() has been
// called on this instance).
type ReaderPool struct {
}
