package store

// A MergeInfo provides information required for a CONTEXT_MERGE context. It is used as part of an IOContext
// in case of CONTEXT_MERGE context.
type MergeInfo struct {
	TotalMaxDoc         int
	EstimatedMergeBytes int
	IsExternal          bool
	MergeMaxNumSegments int
}

func NewMergeInfo(totalMaxDoc int, estimatedMergeBytes int, isExternal bool, mergeMaxNumSegments int) *MergeInfo {
	return &MergeInfo{
		TotalMaxDoc:         totalMaxDoc,
		EstimatedMergeBytes: estimatedMergeBytes,
		IsExternal:          isExternal,
		MergeMaxNumSegments: mergeMaxNumSegments,
	}
}
