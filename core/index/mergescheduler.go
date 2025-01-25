package index

import (
	"io"

	"github.com/geange/lucene-go/core/store"
)

// MergeScheduler
// Expert: IndexWriter uses an instance implementing this interface to execute the merges selected by a MergePolicy.
// The default MergeScheduler is ConcurrentMergeScheduler.
// lucene.experimental
type MergeScheduler interface {
	io.Closer

	// Merge Run the merges provided by MergeScheduler.MergeSource.getNextMerge().
	// Params:
	//		mergeSource – the IndexWriter to obtain the merges from.
	//		trigger – the MergeTrigger that caused this merge to happen
	Merge(mergeSource MergeSource, trigger MergeTrigger) error

	// Initialize IndexWriter calls this on init.
	Initialize(dir store.Directory)
}

type MergeSource interface {
	// GetNextMerge
	// The MergeScheduler calls this method to retrieve the next merge requested by the MergePolicy
	GetNextMerge() (*OneMerge, error)

	// OnMergeFinished
	// Does finishing for a merge.
	OnMergeFinished(merge *OneMerge) error

	// HasPendingMerges
	// Expert: returns true if there are merges waiting to be scheduled.
	HasPendingMerges() bool

	// Merge
	// merges the indicated segments, replacing them in the stack with a single segment.
	Merge(merge *OneMerge) error
}

var _ MergeSource = &indexWriterMergeSource{}

type indexWriterMergeSource struct {
	writer *IndexWriter
}

func (i *indexWriterMergeSource) GetNextMerge() (*OneMerge, error) {
	//TODO implement me
	panic("implement me")
}

func (i *indexWriterMergeSource) OnMergeFinished(merge *OneMerge) error {
	//TODO implement me
	panic("implement me")
}

func (i *indexWriterMergeSource) HasPendingMerges() bool {
	//TODO implement me
	panic("implement me")
}

func (i *indexWriterMergeSource) Merge(merge *OneMerge) error {
	//TODO implement me
	panic("implement me")
}

func newIndexWriterMergeSource(writer *IndexWriter) *indexWriterMergeSource {
	return &indexWriterMergeSource{writer: writer}
}
