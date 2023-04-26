package index

// SortedSetSelector Selects a value from the document's set to use as the representative value
type SortedSetSelector struct {
}

type SortedSetSelectorType int

const (
	MIN        = SortedSetSelectorType(iota) // Selects the minimum value in the set
	MAX                                      // Selects the maximum value in the set
	MIDDLE_MIN                               // Selects the middle value in the set. If the set has an even number of values, the lower of the middle two is chosen.
	MIDDLE_MAX                               // Selects the middle value in the set. If the set has an even number of values, the higher of the middle two is chosen
)
