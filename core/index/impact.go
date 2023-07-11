package index

import "github.com/geange/gods-generic/utils"

// Impact Per-document scoring factors.
type Impact struct {

	// Term frequency of the term in the document.
	Freq int

	// Norm factor of the document.
	Norm int64
}

func NewImpact(freq int, norm int64) *Impact {
	return &Impact{Freq: freq, Norm: norm}
}

func ImpactComparator(c1, c2 *Impact) int {
	//c1 := a.(Impact)
	//c2 := b.(Impact)

	cmp := utils.IntComparator(c1.Freq, c2.Freq)
	if cmp == 0 {
		return utils.Int64Comparator(c1.Norm, c2.Norm)
	}
	return cmp
}
