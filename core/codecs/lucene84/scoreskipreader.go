package lucene84

import (
    "context"
    "github.com/geange/lucene-go/core/interface/index"
    "github.com/geange/lucene-go/core/store"
)

type ScoreSkipReader struct {
    *SkipReader

    impactData       [][]byte
    impactDataLength []int
    badi             *store.ByteArrayDataInput
    impacts          index.Impacts
    numLevels        int
    perLevelImpacts  []*MutableImpactList
}

func (r *ScoreSkipReader) GetImpacts() (index.Impacts, error) {
    panic("")
}

func (r *ScoreSkipReader) SkipTo(ctx context.Context, target int) (int, error) {
    panic("")
}

type MutableImpactList struct {
}
