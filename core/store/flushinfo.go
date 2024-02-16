package store

// A FlushInfo provides information required for a CONTEXT_FLUSH context. It is used as part of an IOContext in
// case of CONTEXT_FLUSH context.
type FlushInfo struct {
	NumDocs              int
	EstimatedSegmentSize int64
}

// NewFlushInfo Creates a new FlushInfo instance from the values required for a CONTEXT_FLUSH IOContext context.
// These values are only estimates and are not the actual values.
func NewFlushInfo(numDocs int, estimatedSegmentSize int64) *FlushInfo {
	return &FlushInfo{NumDocs: numDocs, EstimatedSegmentSize: estimatedSegmentSize}
}
