package index

type Impacts interface {

	// NumLevels Return the number of levels on which we have impacts.
	// The returned value is always greater than 0 and may not always be the same,
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
