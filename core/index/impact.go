package index

import (
	"github.com/geange/gods-generic/utils"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.Impact = &impact{}

// Impact
// Per-document scoring factors.
type impact struct {

	// Term frequency of the term in the document.
	Freq int

	// Norm factor of the document.
	Norm int64
}

func (i *impact) GetFreq() int {
	return i.Freq
}

func (i *impact) GetNorm() int64 {
	return i.Norm
}

func NewImpact(freq int, norm int64) index.Impact {
	return &impact{Freq: freq, Norm: norm}
}

func ImpactComparator(c1, c2 index.Impact) int {
	//c1 := a.(Impact)
	//c2 := b.(Impact)

	cmp := utils.IntComparator(c1.GetFreq(), c2.GetFreq())
	if cmp == 0 {
		return utils.Int64Comparator(c1.GetNorm(), c2.GetNorm())
	}
	return cmp
}
