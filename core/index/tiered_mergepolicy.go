package index

const (
	// TIERED_MERGE_POLICY_DEFAULT_NO_CFS_RATIO
	// Default noCFSRatio. If a merge's size is >= 10% of the index, then we disable compound file for it.
	// See Also: MergePolicy.setNoCFSRatio
	TIERED_MERGE_POLICY_DEFAULT_NO_CFS_RATIO = 0.1
)

var _ MergePolicy = &TieredMergePolicy{}

type TieredMergePolicy struct {
	*MergePolicyDefault

	// User-specified maxMergeAtOnce. In practice we always take the min of its
	// value and segsPerTier to avoid suboptimal merging.
	maxMergeAtOnce              int
	maxMergedSegmentBytes       int64
	maxMergeAtOnceExplicit      int
	floorSegmentBytes           int64
	segsPerTier                 float64
	forceMergeDeletesPctAllowed float64
	deletesPctAllowed           float64
}

func (t *TieredMergePolicy) FindMerges(mergeTrigger MergeTrigger, segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TieredMergePolicy) FindForcedMerges(segmentInfos *SegmentInfos, maxSegmentCount int, segmentsToMerge map[*SegmentCommitInfo]bool, mergeContext MergeContext) (*MergeSpecification, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TieredMergePolicy) FindForcedDeletesMerges(segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TieredMergePolicy) FindFullFlushMerges(mergeTrigger MergeTrigger, segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TieredMergePolicy) UseCompoundFile(infos *SegmentInfos, mergedInfo *SegmentCommitInfo, mergeContext MergeContext) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TieredMergePolicy) Size(info *SegmentCommitInfo, mergeContext MergeContext) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TieredMergePolicy) GetNoCFSRatio() float64 {
	//TODO implement me
	panic("implement me")
}
