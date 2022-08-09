package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

const (
	BLOCK_SIZE = 8
)

type SkipWriter struct {
}

func (s *SkipWriter) resetSkip() {

}

func (s *SkipWriter) bufferSkip(doc int, pointer int64, count int, accumulator *index.CompetitiveImpactAccumulator) {

}

func (s *SkipWriter) writeSkip(out store.IndexOutput) {

}
