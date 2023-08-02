package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"slices"
)

// OrdinalMap
// Maps per-segment ordinals to/from global ordinal space, using a compact packed-ints representation.
// NOTE: this is a costly operation, as it must merge sort all terms, and may require non-trivial RAM once done.
// It's better to operate in segment-private ordinal space instead when possible.
// lucene.internal
type OrdinalMap struct {
	valueCount          int64 // number of global ordinals
	globalOrdDeltas     types.LongValues
	firstSegments       types.LongValues
	segmentToGlobalOrds []types.LongValues
	segmentMap          *SegmentMap
}

func NewOrdinalMap(subs []index.TermsEnum, segmentMap *SegmentMap, acceptableOverheadRatio float64) (*OrdinalMap, error) {
	//res := &OrdinalMap{
	//	segmentMap: segmentMap,
	//}
	//globalOrdDeltas := packed.NewPackedLongValuesBuilder()
	//firstSegments := packed.NewPackedLongValuesBuilder()
	//ordDeltas := make([]*packed.LongValuesBuilder, len(subs))
	//firstSegmentBits := 0
	//for i := 0; i < len(ordDeltas); i++ {
	//	ordDeltas[i] = packed.
	//}
	panic("")
}

type TermsEnumIndex struct {
	subIndex    int
	termsEnum   index.TermsEnum
	currentTerm []byte
}

func NewTermsEnumIndex(termsEnum index.TermsEnum, subIndex int) *TermsEnumIndex {
	return &TermsEnumIndex{
		subIndex:  subIndex,
		termsEnum: termsEnum,
	}
}

func (t *TermsEnumIndex) Next(ctx context.Context) ([]byte, error) {
	next, err := t.termsEnum.Next(ctx)
	if err != nil {
		return nil, err
	}
	t.currentTerm = next
	return next, nil
}

type SegmentMap struct {
	newToOld, oldToNew []int
}

func (s *SegmentMap) NewToOld(segment int) int {
	return s.newToOld[segment]
}

func (s *SegmentMap) OldToNew(segment int) int {
	return s.oldToNew[segment]
}

func NewSegmentMap(weight []int64) *SegmentMap {
	newToOld := makeIndexMap(weight)
	oldToNew := inverseInts(newToOld)
	return &SegmentMap{
		newToOld: newToOld,
		oldToNew: oldToNew,
	}
}

func makeIndexMap(weight []int64) []int {
	newToOld := make([]int, len(weight))
	for i := 0; i < len(weight); i++ {
		newToOld[i] = i
	}

	slices.SortFunc(newToOld, func(i, j int) int {
		return Compare(weight[newToOld[j]], weight[newToOld[i]])
	})
	return newToOld
}

func inverseInts(data []int) []int {
	inverse := slices.Clone(data)
	slices.Reverse(inverse)
	return inverse
}
