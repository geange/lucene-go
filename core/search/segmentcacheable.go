package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

// SegmentCacheable
// Interface defining whether or not an object can be cached against a LeafReader Objects
// that depend only on segment-immutable structures such as Points or postings lists can
// just return true from isCacheable(LeafReaderContext) Objects that depend on doc values
// should return DocValues.isCacheable(LeafReaderContext, String...), which will check to
// see if the doc values fields have been updated. Updated doc values fields are not suitable
// for cacheing. Objects that are not segment-immutable, such as those that rely on global
// statistics or scores, should return false
type SegmentCacheable interface {

	// IsCacheable
	// Returns: true if the object can be cached against a given leaf
	IsCacheable(ctx index.LeafReaderContext) bool
}
