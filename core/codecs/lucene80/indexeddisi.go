package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// jump-table time/space trade-offs to consider:
// The block offsets and the block indexes could be stored in more compressed form with
// two PackedInts or two MonotonicDirectReaders.
// The DENSE ranks (default 128 shorts = 256 bytes) could likewise be compressed. But as there is
// at least 4096 set bits in DENSE blocks, there will be at least one rank with 2^12 bits, so it
// is doubtful if there is much to gain here.

const (
	BLOCK_SIZE               = 65536
	DENSE_BLOCK_LONGS        = BLOCK_SIZE / 64
	DEFAULT_DENSE_RANK_POWER = 9
	MAX_ARRAY_LENGTH         = (1 << 12) - 1
)

var _ types.DocIdSetIterator = &IndexedDISI{}

type IndexedDISI struct {
	slice               store.IndexInput
	jumpTableEntryCount int
	denseRankPower      byte
	jumpTable           store.RandomAccessInput // Skip blocks of 64K bits
	denseRankTable      []byte
	cost                int64

	block             int
	blockEnd          int64
	denseBitmapOffset int64
	nextBlockIndex    int
	method            Method

	doc   int
	index int

	// SPARSE variables
	exists bool

	// DENSE variables
	word      int64
	wordIndex int
	// number of one bits encountered so far, including those of `word`
	numberOfOnes int
	// Used with rank for jumps inside of DENSE as they are absolute instead of relative
	denseOrigoIndex int

	// ALL variables
	gap int
}

func (i *IndexedDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) AdvanceExact(ctx context.Context, target int) (bool, error) {
	panic("implement me")
}

type Method struct{}
