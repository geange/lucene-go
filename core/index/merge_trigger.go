package index

type MergeTrigger int

const (
	// SEGMENT_FLUSH Merge was triggered by a segment flush.
	SEGMENT_FLUSH = MergeTrigger(iota)

	// FULL_FLUSH Merge was triggered by a full flush. Full flushes can be caused by a commit,
	// NRT reader reopen or a close call on the index writer.
	FULL_FLUSH

	// EXPLICIT Merge has been triggered explicitly by the user.
	EXPLICIT

	// MERGE_FINISHED Merge was triggered by a successfully finished merge.
	MERGE_FINISHED

	// CLOSING Merge was triggered by a closing IndexWriter.
	CLOSING

	// COMMIT Merge was triggered on commit.
	COMMIT

	// GET_READER Merge was triggered on opening NRT readers.
	GET_READER
)
