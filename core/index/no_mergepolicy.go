package index

import "math"

var _ MergePolicy = &NoMergePolicy{}

type NoMergePolicy struct {
	*MergePolicyBase
}

func NewNoMergePolicy() *NoMergePolicy {
	policy := &NoMergePolicy{}
	policy.MergePolicyBase = NewMergePolicy(policy)
	return policy
}

func (n *NoMergePolicy) FindMerges(mergeTrigger MergeTrigger, segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	return nil, nil
}

func (n *NoMergePolicy) FindForcedMerges(segmentInfos *SegmentInfos, maxSegmentCount int, segmentsToMerge map[*SegmentCommitInfo]bool, mergeContext MergeContext) (*MergeSpecification, error) {
	return nil, nil
}

func (n *NoMergePolicy) FindForcedDeletesMerges(segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	return nil, nil
}

func (n *NoMergePolicy) FindFullFlushMerges(mergeTrigger MergeTrigger, segmentInfos *SegmentInfos, mergeContext MergeContext) (*MergeSpecification, error) {
	return nil, nil
}

func (n *NoMergePolicy) UseCompoundFile(infos *SegmentInfos, newSegment *SegmentCommitInfo, mergeContext MergeContext) (bool, error) {
	return newSegment.info.GetUseCompoundFile(), nil
}

func (n *NoMergePolicy) Size(info *SegmentCommitInfo, mergeContext MergeContext) (int64, error) {
	return math.MaxInt64, nil
}

func (n *NoMergePolicy) GetNoCFSRatio() float64 {
	return n.MergePolicyBase.getNoCFSRatio()
}
