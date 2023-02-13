package packed

import "github.com/geange/lucene-go/core/types"

var _ types.LongValues = &PackedLongValues{}

// PackedLongValues Utility class to compress integers into a LongValues instance.
type PackedLongValues struct {
	values              []*PackedIntsReader
	pageShift, pageMask int
	size                int64
}

func (p *PackedLongValues) Get(index int64) int64 {
	//TODO implement me
	panic("implement me")
}
