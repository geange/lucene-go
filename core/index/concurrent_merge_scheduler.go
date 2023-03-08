package index

import (
	"github.com/geange/lucene-go/core/store"
)

var _ MergeScheduler = &ConcurrentMergeScheduler{}

// ConcurrentMergeScheduler A MergeScheduler that runs each merge using a separate thread.
// Specify the max number of threads that may run at once, and the maximum number of
// simultaneous merges with setMaxMergesAndThreads.
//
// If the number of merges exceeds the max number of threads then the largest merges are
// paused until one of the smaller merges completes.
//
// If more than getMaxMergeCount merges are requested then this class will forcefully throttle
// the incoming threads by pausing until one more merges complete.
//
// This class attempts to detect whether the index is on rotational storage (traditional hard drive)
// or not (e.g. solid-state disk) and changes the default max merge and thread count accordingly.
// This detection is currently Linux-only, and relies on the OS to put the right value into
// /sys/block/<dev>/block/rotational. For all other operating systems it currently assumes a
// rotational disk for backwards compatibility. To enable default settings for spinning or
// solid state disks for such operating systems, use setDefaultMaxMergesAndThreads(boolean).
type ConcurrentMergeScheduler struct {
}

func (c *ConcurrentMergeScheduler) Merge(mergeSource MergeSource, trigger MergeTrigger) error {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentMergeScheduler) Close() error {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentMergeScheduler) Initialize(dir store.Directory) {
	//TODO implement me
	panic("implement me")
}
