package index

import "io"

// MergeScheduler Expert: IndexWriter uses an instance implementing this interface to execute the merges selected by a MergePolicy. The default MergeScheduler is ConcurrentMergeScheduler.
// lucene.experimental
type MergeScheduler interface {
	// Merge Run the merges provided by MergeScheduler.MergeSource.getNextMerge().
	// Params:
	//		mergeSource – the IndexWriter to obtain the merges from.
	//		trigger – the MergeTrigger that caused this merge to happen
	Merge(mergeSource MergeSource, trigger MergeTrigger) error

	io.Closer
}

type MergeSource interface {
}
