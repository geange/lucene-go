package index

type MergeTrigger int

const (
	// MERGE_TRIGGER_SEGMENT_FLUSH
	// Merge was triggered by a segment Flush.
	// 由一个段的flush来触发的
	MERGE_TRIGGER_SEGMENT_FLUSH = MergeTrigger(iota)

	// MERGE_TRIGGER_FULL_FLUSH
	// Merge was triggered by a full Flush. Full flushes can be caused by a commit,
	// NRT reader reopen or a close call on the index writer.
	MERGE_TRIGGER_FULL_FLUSH

	// MERGE_TRIGGER_EXPLICIT
	// Merge has been triggered explicitly by the user.
	MERGE_TRIGGER_EXPLICIT

	// MERGE_TRIGGER_MERGE_FINISHED
	// Merge was triggered by a successfully finished merge.
	MERGE_TRIGGER_MERGE_FINISHED

	// MERGE_TRIGGER_CLOSING
	// Merge was triggered by a closing IndexWriter.
	MERGE_TRIGGER_CLOSING

	// MERGE_TRIGGER_COMMIT
	// Merge was triggered on commit.
	MERGE_TRIGGER_COMMIT

	// MERGE_TRIGGER_GET_READER
	// Merge was triggered on opening NRT readers.
	MERGE_TRIGGER_GET_READER
)
