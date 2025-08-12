package lucene50

import (
	"github.com/geange/lucene-go/core/codecs/compressing"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.TermVectorsFormat = &TermVectorsFormat{}

type TermVectorsFormat struct {
	*compressing.TermVectorsFormat
}

func NewTermVectorsFormat() *TermVectorsFormat {
	return &TermVectorsFormat{
		compressing.NewTermVectorsFormat("Lucene50TermVectorsData", "", compressing.FAST, 1<<12, 128, 10),
	}
}
