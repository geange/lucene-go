package index

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/util/bytesutils"
	"math"
)

// FieldUpdatesBuffer This class efficiently buffers numeric and binary field updates and stores terms,
// values and metadata in a memory efficient way without creating large amounts of objects. Update
// terms are stored without de-duplicating the update term. In general we try to optimize for several
// use-cases. For instance we try to use constant space for update terms field since the common case always
// updates on the same field. Also for docUpTo we try to optimize for the case when updates should be applied
// to all docs ie. docUpTo=Integer.MAX_VALUE. In other cases each update will likely have a different docUpTo.
// Along the same lines this impl optimizes the case when all updates have a item. Lastly, if all updates
// share the same item for a numeric field we only store the item once.
type FieldUpdatesBuffer struct {
	numUpdates    int
	termValues    *bytesutils.BytesRefArray
	termSortState *bytesutils.SortState
	byteValues    *bytesutils.BytesRefArray
	docsUpTo      []int
	numericValues []int64
	hasValues     *bitset.BitSet
	maxNumeric    int64
	minNumeric    int64
	fields        []string
	isNumeric     bool
	finished      bool
}

func NewDefaultFieldUpdatesBuffer() *FieldUpdatesBuffer {
	return &FieldUpdatesBuffer{
		numUpdates:    1,
		termValues:    nil,
		termSortState: nil,
		byteValues:    nil,
		docsUpTo:      nil,
		numericValues: nil,
		hasValues:     nil,
		maxNumeric:    math.MaxInt64,
		minNumeric:    math.MinInt64,
		fields:        nil,
		isNumeric:     false,
		finished:      false,
	}
}
