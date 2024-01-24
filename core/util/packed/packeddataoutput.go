package packed

import "github.com/geange/lucene-go/core/store"

// PackedDataOutput
// A DataOutput wrapper to write unaligned, variable-length packed integers.
// 请参阅: PackedDataInput
// lucene.internal
type PackedDataOutput struct {
	out           store.DataOutput
	current       int64
	remainingBits int
}
