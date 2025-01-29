package analysis

import (
	"unicode"
)

type WhitespaceTokenizer struct {
	CharTokenizerBase
}

func (w *WhitespaceTokenizer) IsTokenChar(r rune) bool {
	return !unicode.IsSpace(r)
}
