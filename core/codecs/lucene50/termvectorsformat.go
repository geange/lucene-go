package lucene50

import (
	"github.com/geange/lucene-go/core/codecs/compressing"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.TermVectorsFormat = &TermVectorsFormat{}

type TermVectorsFormat struct {
	*compressing.TermVectorsFormat
}
