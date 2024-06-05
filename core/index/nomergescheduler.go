package index

import "github.com/geange/lucene-go/core/store"

var _ MergeScheduler = &NoMergeScheduler{}

type NoMergeScheduler struct {
}

func NewNoMergeScheduler() *NoMergeScheduler {
	return &NoMergeScheduler{}
}

func (n *NoMergeScheduler) Close() error {
	return nil
}

func (n *NoMergeScheduler) Merge(mergeSource MergeSource, trigger MergeTrigger) error {
	return nil
}

func (n *NoMergeScheduler) Initialize(dir store.Directory) {
	return
}
