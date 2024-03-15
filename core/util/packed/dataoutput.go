package packed

import "github.com/geange/lucene-go/core/store"

// DataOutput
// A DataOutput wrapper to write unaligned, variable-length packed integers.
// 请参阅: DataInput
// lucene.internal
type DataOutput struct {
	out           store.DataOutput
	current       int64
	remainingBits int
}
