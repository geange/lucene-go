package index

import "github.com/geange/gods-generic/utils"

// Impact
// Per-document scoring factors.
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

type Impacts interface {

	// NumLevels Return the number of levels on which we have impacts.
	// The returned item is always greater than 0 and may not always be the same,
	// even on a single postings list, depending on the current doc ID.
	NumLevels() int

	// GetDocIdUpTo Return the maximum inclusive doc ID until which the list of impacts
	// returned by getImpacts(int) is valid. This is a non-decreasing function of level.
	GetDocIdUpTo(level int) int

	// GetImpacts Return impacts on the given level. These impacts are sorted by increasing
	// frequency and increasing unsigned norm, and only valid until the doc ID returned by
	// getDocIdUpTo(int) for the same level, included. The returned list is never empty.
	// NOTE: There is no guarantee that these impacts actually appear in postings, only that
	// they trigger scores that are greater than or equal to the impacts that actually
	// appear in postings.
	GetImpacts(level int) []*Impact
}

// ImpactsEnum Extension of PostingsEnum which also provides information about upcoming impacts.
type ImpactsEnum interface {
	PostingsEnum
	ImpactsSource
}

// ImpactsSource Source of Impacts.
type ImpactsSource interface {
	// AdvanceShallow Shallow-advance to target. This is cheaper than calling DocIdSetIterator.advance(int)
	// and allows further calls to getImpacts() to ignore doc IDs that are less than target in order to get
	// more precise information about impacts. This method may not be called on targets that are less than
	// the current DocIdSetIterator.docID(). After this method has been called, DocIdSetIterator.nextDoc()
	// may not be called if the current doc ID is less than target - 1 and DocIdSetIterator.advance(int)
	// may not be called on targets that are less than target.
	AdvanceShallow(target int) error

	// GetImpacts Get information about upcoming impacts for doc ids that are greater than or equal to the
	// maximum of DocIdSetIterator.docID() and the last target that was passed to advanceShallow(int).
	// This method may not be called on an unpositioned iterator on which advanceShallow(int) has never been
	// called. NOTE: advancing this iterator may invalidate the returned impacts, so they should not be used
	// after the iterator has been advanced.
	GetImpacts() (Impacts, error)
}
