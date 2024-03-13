package packed

import "github.com/geange/lucene-go/core/store"

// DataInput
// A DataInput wrapper to read unaligned, variable-length packed integers.
// This API is much slower than the PackedInts fixed-length API but can be convenient to save space.
// 请参阅: DataOutput
// lucene.internal
type DataInput struct {
	in            store.DataInput
	current       int64
	remainingBits int
}
