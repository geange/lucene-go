package index

// CompetitiveImpactAccumulator This class accumulates the (freq, norm) pairs that may produce competitive scores.
type CompetitiveImpactAccumulator struct {
}

func (c *CompetitiveImpactAccumulator) Clear() {

}

// Add Accumulate a (freq,norm) pair, updating this structure if there is no equivalent or more competitive entry already.
func (c *CompetitiveImpactAccumulator) Add(freq int, norm int64) {

}
