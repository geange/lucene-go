package standard

import (
	"github.com/geange/lucene-go/core/analysis"
	"io"
)

type StandardTokenizer struct {
	*analysis.TokenizerImp

	scanner *StandardTokenizerImpl

	skippedPositions int
	maxTokenLength   int
}

func NewTokenizer() *StandardTokenizer {
	tokenizer := &StandardTokenizer{
		TokenizerImp:     analysis.NewTokenizerImpl(),
		scanner:          NewStandardTokenizerImpl(),
		skippedPositions: 0,
		maxTokenLength:   0,
	}
	return tokenizer
}

func (r *StandardTokenizer) IncrementToken() (bool, error) {
	r.AttributeSource().Clear()
	r.skippedPositions = 0

	text, err := r.scanner.GetNextToken()
	if err != nil {
		return false, err
	}

	r.AttributeSource().CharTerm().Append(text)
	r.AttributeSource().Offset().SetOffset(r.scanner.Slow, r.scanner.Slow+len(text))
	return true, nil
}

func (r *StandardTokenizer) SetReader(reader io.Reader) error {
	r.scanner.SetReader(reader)
	return nil
}

func (r *StandardTokenizer) setMaxTokenLength(length int) {
	r.maxTokenLength = length
}
