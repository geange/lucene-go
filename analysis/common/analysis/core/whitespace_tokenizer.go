package core

import (
	"github.com/geange/lucene-go/analysis/common/analysis/util"
	"unicode"
)

type WhitespaceTokenizer struct {
	util.CharTokenizerBase
}

func (w *WhitespaceTokenizer) IsTokenChar(r rune) bool {
	return !unicode.IsSpace(r)
}
