package search

import (
	"github.com/geange/gods-generic/utils"
	"go.uber.org/atomic"
	"math"
)

const (
	// DEFAULT_INTERVAL
	// we use 2^10-1 to check the remainder with a bitwise operation
	DEFAULT_INTERVAL = 0x3ff
)

// MaxScoreAccumulator
// Maintains the maximum score and its corresponding document id concurrently
type MaxScoreAccumulator struct {
	acc         *atomic.Int64
	modInterval int64
}

func NewMaxScoreAccumulator() *MaxScoreAccumulator {
	return &MaxScoreAccumulator{
		acc:         atomic.NewInt64(math.MinInt64),
		modInterval: DEFAULT_INTERVAL,
	}
}

func (m *MaxScoreAccumulator) Accumulate(docBase int, score float32) error {
	// encode = (((long) Float.floatToIntBits(score)) << 32) | docBase
	encode := int64(math.Float32bits(score))<<32 | int64(docBase)
	m.acc.Store(m.maxEncode(m.acc.Load(), encode))
	return nil
}

// Return the max encoded DocAndScore in a way that is consistent with MaxScoreAccumulator.DocAndScore.compareTo.
func (m *MaxScoreAccumulator) maxEncode(v1, v2 int64) int64 {
	score1 := math.Float32frombits(uint32(v1 >> 32))
	score2 := math.Float32frombits(uint32(v2 >> 32))

	cmp := utils.Float32Comparator(score1, score2)
	if cmp == 0 {
		// tie-break on the minimum doc base
		if v1 < v2 {
			return v1
		}
		return v2
	}

	if cmp > 0 {
		return v1
	}
	return v2
}

func (m *MaxScoreAccumulator) Get() *DocAndScore {
	value := m.acc.Load()
	if value == math.MinInt64 {
		return nil
	}

	score := math.Float64frombits(uint64(value >> 32))
	docBase := int(value)
	return NewDocAndScore(docBase, score)
}

type DocAndScore struct {
	docBase int
	score   float64
}

func NewDocAndScore(docBase int, score float64) *DocAndScore {
	return &DocAndScore{
		docBase: docBase,
		score:   score,
	}
}
