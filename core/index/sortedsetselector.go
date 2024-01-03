package index

// SortedSetSelector Selects a item from the document's set to use as the representative item
type SortedSetSelector struct {
}

type SortedSetSelectorType int

const (
	MIN        = SortedSetSelectorType(iota) // Selects the minimum item in the set
	MAX                                      // Selects the maximum item in the set
	MIDDLE_MIN                               // Selects the middle item in the set. If the set has an even number of values, the lower of the middle two is chosen.
	MIDDLE_MAX                               // Selects the middle item in the set. If the set has an even number of values, the higher of the middle two is chosen
)
