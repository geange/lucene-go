package index

import (
	"math"
	"sync"
)

const (
	DEFAULT_NO_CFS_RATIO         = 1.0
	DEFAULT_MAX_CFS_SEGMENT_SIZE = math.MaxInt64
)

// MergePolicy
// Expert: a MergePolicy determines the sequence of primitive merge operations.
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

	// FindMerges
	// Determine what set of merge operations are now necessary on the index.
	// IndexWriter calls this whenever there is a change to the segments.
	// This call is always synchronized on the IndexWriter instance so only one thread at a time will call this method.
	// mergeTrigger: the event that triggered the merge
	// segmentInfos: the total set of segments in the index
	// mergeContext: the IndexWriter to find the merges on
	FindMerges(mergeTrigger MergeTrigger, segmentInfos *SegmentInfos,
		mergeContext MergeContext) (*MergeSpecification, error)

	// FindForcedMerges Determine what set of merge operations is necessary in order to
	// merge to <= the specified segment count.
	// IndexWriter calls this when its IndexWriter.forceMerge method is called.
	// This call is always synchronized on the IndexWriter instance so only one
	// thread at a time will call this method.
	//
	// FindForcedMerges确定需要哪一组合并操作才能合并到<=指定的段计数。
	// IndexWriter在调用其IndexWriter.forceMerge方法时调用此函数。
	// 此调用始终在IndexWriter实例上同步，因此一次只有一个线程会调用此方法
	//
	// Params:
	//			segmentInfos – the total set of segments in the index
	//	 		maxSegmentCount – requested maximum number of segments in the index (currently this is always 1)
	//	 		segmentsToMerge – contains the specific SegmentInfo instances that must be merged away.
	//	 			This may be a subset of all SegmentInfos. If the item is True for a given SegmentInfo,
	//	 			that means this segment was an original segment present in the to-be-merged index;
	//	 			else, it was a segment produced by a cascaded merge.
	//	 		mergeContext – the MergeContext to find the merges on
	FindForcedMerges(segmentInfos *SegmentInfos, maxSegmentCount int,
		segmentsToMerge map[*SegmentCommitInfo]bool, mergeContext MergeContext) (*MergeSpecification, error)

	// FindForcedDeletesMerges
	// Determine what set of merge operations is necessary in order to expunge all deletes from the index.
	//
	// 确定需要哪一组合并操作才能从索引中删除所有删除。
	//
	// Params:
	//			segmentInfos – the total set of segments in the index
	//			mergeContext – the MergeContext to find the merges on
	FindForcedDeletesMerges(segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error)

	// FindFullFlushMerges
	// Identifies merges that we want to execute (synchronously) on commit.
	// By default, this will do no merging on commit. If you implement this method in your MergePolicy you
	// must also set a non-zero timeout using IndexWriterConfig.setMaxFullFlushMergeWaitMillis.
	// Any merges returned here will make IndexWriter.commit(), IndexWriter.prepareCommit() or
	// IndexWriter.getReader(boolean, boolean) block until the merges complete or until
	// IndexWriterConfig.getMaxFullFlushMergeWaitMillis() has elapsed.
	// This may be used to merge small segments that have just been flushed, reducing the number of
	// segments in the point in time snapshot. If a merge does not complete in the allotted time,
	// it will continue to execute, and eventually finish and apply to future point in time snapshot,
	// but will not be reflected in the current one. If a MergePolicy.OneMerge in the returned
	// MergePolicy.MergeSpecification includes a segment already included in a registered merge,
	// then IndexWriter.commit() or IndexWriter.prepareCommit() will throw a IllegalStateException.
	// Use MergePolicy.MergeContext.getMergingSegments() to determine which segments are currently
	// registered to merge.
	//
	// 标识我们要在提交时（同步）执行的合并。默认情况下，这将不会在提交时进行合并。
	// 如果在MergePolicy中实现此方法，则还必须使用IndexWriterConfig.setMaxFullFlushMergeWaitMillis设置非零超时。
	// 此处返回的任何合并都将导致IndexWriter.commit（）、IndexWriter.prepareCommit()
	// 或IndexWriter.getReader（布尔值、布尔值）阻塞，直到合并完成或IndexWriter
	// Config.getMaxFullFlushMergeWaitMillis()结束。这可以用于合并刚刚刷新的小片段，
	// 从而减少时间点快照中的片段数量。如果合并没有在分配的时间内完成，它将继续执行，
	// 最终完成并应用于未来的时间点快照，但不会反映在当前快照中。如果返回的
	// MergePolicy.MergeSpecification中的MergePolicy.OneMerge包含已包含在注册合并中的段，则IndexWriter.commit()
	// 或IndexWriter.prepareCommit（）将引发IllegalStateException。
	// 使用MergePolicy.MergeContext.getMergingSegments()确定当前要注册合并的段。
	//
	// Params:
	//			mergeTrigger – the event that triggered the merge (COMMIT or GET_READER).
	//			segmentInfos – the total set of segments in the index (while preparing the commit)
	//			mergeContext – the MergeContext to find the merges on, which should be used to
	//				determine which segments are already in a registered merge
	//				(see MergePolicy.MergeContext.getMergingSegments()).
	FindFullFlushMerges(mergeTrigger MergeTrigger,
		segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error)

	// UseCompoundFile
	// Returns true if a new segment (regardless of its origin) should use the compound file format.
	// The default implementation returns true iff the size of the given mergedInfo is less or equal
	// to getMaxCFSSegmentSizeMB() and the size is less or equal to the TotalIndexSize * getNoCFSRatio()
	// otherwise false.
	//
	// 如果新段（无论其来源如何）应使用复合文件格式，则返回true。
	// 如果给定mergedInfo的大小小于或等于getMaxCFSSegmentSizeMB()，
	// 并且大小小于或相等于TotalIndexSize*getNoCFSRatio()，则默认实现返回true，否则为false。
	UseCompoundFile(infos *SegmentInfos,
		mergedInfo *SegmentCommitInfo, mergeContext MergeContext) (bool, error)

	KeepFullyDeletedSegment(func() CodecReader) bool

	MergePolicyInner
}

type MergePolicyInner interface {
	Size(info *SegmentCommitInfo, mergeContext MergeContext) (int64, error)
	GetNoCFSRatio() float64
}

func NewMergePolicy(inner MergePolicyInner) *MergePolicyBase {
	return &MergePolicyBase{
		MergePolicyInner:  inner,
		noCFSRatio:        DEFAULT_NO_CFS_RATIO,
		maxCFSSegmentSize: DEFAULT_MAX_CFS_SEGMENT_SIZE,
	}
}

type MergePolicyBase struct {
	MergePolicyInner

	noCFSRatio        float64
	maxCFSSegmentSize int64
}

func (m *MergePolicyBase) FindFullFlushMerges(mergeTrigger MergeTrigger,
	segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	return nil, nil
}

func (m *MergePolicyBase) UseCompoundFile(infos *SegmentInfos,
	mergedInfo *SegmentCommitInfo, mergeContext MergeContext) (bool, error) {

	if m.GetNoCFSRatio() == 0.0 {
		return false, nil
	}
	mergedInfoSize, err := m.Size(mergedInfo, mergeContext)
	if err != nil {
		return false, err
	}
	if mergedInfoSize > m.maxCFSSegmentSize {
		return false, nil
	}
	if m.GetNoCFSRatio() >= 1.0 {
		return true, nil
	}
	totalSize := int64(0)
	for _, info := range infos.AsList() {
		size, _ := m.Size(info, mergeContext)
		totalSize += size
	}
	return float64(mergedInfoSize) <= m.GetNoCFSRatio()*float64(totalSize), nil
}

func (m *MergePolicyBase) size(info *SegmentCommitInfo, mergeContext MergeContext) (int64, error) {
	byteSize, err := info.SizeInBytes()
	if err != nil {
		return 0, err
	}
	delCount, err := mergeContext.NumDeletesToMerge(info)
	if err != nil {
		return 0, err
	}

	maxDoc, err := info.info.MaxDoc()
	if err != nil {
		return 0, err
	}

	delRatio := float64(0)
	if maxDoc > 0 {
		delRatio = float64(delCount) / float64(maxDoc)
	}

	if maxDoc <= 0 {
		return byteSize, nil
	}
	return int64(float64(byteSize) * (1.0 - delRatio)), nil
}

func (m *MergePolicyBase) getNoCFSRatio() float64 {
	return m.noCFSRatio
}

func (m *MergePolicyBase) KeepFullyDeletedSegment(func() CodecReader) bool {
	return false
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
	info           *SegmentCommitInfo // used by IndexWriter
	registerDone   bool               // used by IndexWriter
	mergeGen       bool               // used by IndexWriter
	isExternal     bool               // used by IndexWriter
	maxNumSegments int                // used by IndexWriter
}

// A MergeSpecification instance provides the information necessary to perform multiple merges.
// It simply contains a list of MergePolicy.OneMerge instances.
type MergeSpecification struct {
	// The subset of segments to be included in the primitive merge.
	merges []*OneMerge
}

// NewMergeSpecification Sole constructor. Use add(MergePolicy.OneMerge) to add merges.
func NewMergeSpecification() *MergeSpecification {
	return &MergeSpecification{merges: make([]*OneMerge, 0)}
}

func (m *MergeSpecification) Add(merge *OneMerge) {
	m.merges = append(m.merges, merge)
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
