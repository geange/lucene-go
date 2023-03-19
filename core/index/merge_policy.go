package index

import "sync"

// MergePolicy Expert: a MergePolicy determines the sequence of primitive merge operations.
// Whenever the segments in an index have been altered by IndexWriter, either the addition of a newly
// flushed segment, addition of many segments from addIndexes* calls, or a previous merge that may now
// need to cascade, IndexWriter invokes findMerges to give the MergePolicy a chance to pick merges that
// are now required. This method returns a MergePolicy.MergeSpecification instance describing the set
// of merges that should be done, or null if no merges are necessary. When IndexWriter.forceMerge is
// called, it calls findForcedMerges(SegmentInfos, int, Map, MergePolicy.MergeContext) and the MergePolicy
// should then return the necessary merges.
//
// Note that the policy can return more than one merge at a time. In this case, if the writer is using
// SerialMergeScheduler, the merges will be run sequentially but if it is using ConcurrentMergeScheduler
// they will be run concurrently.
//
// The default MergePolicy is TieredMergePolicy.
//
// lucene.experimental
type MergePolicy interface {
}

// MergeContext This interface represents the current context of the merge selection process. It allows
// to access real-time information like the currently merging segments or how many deletes a segment
// would claim back if merged. This context might be stateful and change during the execution of a
// merge policy's selection processes.
// lucene.experimental
type MergeContext interface {

	// NumDeletesToMerge Returns the number of deletes a merge would claim back if the given segment is merged.
	// Params: info – the segment to get the number of deletes for
	// See Also: numDeletesToMerge(SegmentCommitInfo, int, IOSupplier)
	NumDeletesToMerge(info *SegmentCommitInfo) (int, error)

	// NumDeletedDocs Returns the number of deleted documents in the given segments.
	NumDeletedDocs(info *SegmentCommitInfo) int

	// Returns the info stream that can be used to log messages
	//getInfoStream() util.InfoStream

	// GetMergingSegments Returns an unmodifiable set of segments that are currently merging.
	GetMergingSegments() []*SegmentCommitInfo
}

// OneMerge provides the information necessary to perform an individual primitive merge operation,
// resulting in a single new segment. The merge spec includes the subset of segments to be merged
// as well as whether the new segment should use the compound file format.
// lucene.experimental
type OneMerge struct {
	info           *SegmentCommitInfo
	registerDone   bool
	mergeGen       bool
	isExternal     bool
	maxNumSegments int
}

type MergeSpecification struct {
	merges []*OneMerge
}

// OneMergeProgress Progress and state for an executing merge. This class encapsulates the logic to pause
// and resume the merge thread or to abort the merge entirely.
// lucene.experimental
type OneMergeProgress struct {
	pauseLock sync.Mutex
}

// PauseReason Reason for pausing the merge thread.
type PauseReason int

const (
	STOPPED = PauseReason(iota) // Stopped (because of throughput rate set to 0, typically).
	PAUSED                      // Temporarily paused because of exceeded throughput rate.
	OTHER                       // Other reason.
)
