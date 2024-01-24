package packed

import "github.com/geange/lucene-go/core/store"

type AbstractBlockPackedWriter struct {
	out      store.DataOutput
	values   []int64
	blocks   []byte
	ord      int64
	finished bool

	MIN_BLOCK_SIZE     int
	MAX_BLOCK_SIZE     int
	MIN_VALUE_EQUALS_0 int
	BPV_SHIFT          int
}

func (a *AbstractBlockPackedWriter) initConst() {
	a.MIN_BLOCK_SIZE = 64
	a.MAX_BLOCK_SIZE = 1 << (30 - 3)
	a.MIN_VALUE_EQUALS_0 = 1 << 0
	a.BPV_SHIFT = 1
}

// Reset this writer to wrap out. The block size remains unchanged.
func (a *AbstractBlockPackedWriter) reset(out store.DataOutput) {
	a.out = out
	a.values = a.values[:0]
	a.ord = 0
	a.finished = false
}

func (a *AbstractBlockPackedWriter) flush() error {
	panic("")
}
