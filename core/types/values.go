package types

// LongValues Abstraction over an array of longs.
// lucene.internal
type LongValues interface {
	Get(index int64) int64
}
