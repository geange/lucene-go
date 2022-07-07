package standard

import (
	"github.com/geange/lucene-go/core/analysis"
	"io"
)

type Tokenizer struct {
	*analysis.TokenizerImp

	scanner *TokenizerImpl

	skippedPositions int
	maxTokenLength   int
}

func NewTokenizer(reader io.Reader) *Tokenizer {
	tokenizer := &Tokenizer{
		TokenizerImp:     analysis.NewTokenizerImpl(),
		scanner:          &TokenizerImpl{},
		skippedPositions: 0,
		maxTokenLength:   0,
	}
	tokenizer.SetReader(reader)
	tokenizer.Input = reader
	return tokenizer
}

func (r *Tokenizer) IncrementToken() (bool, error) {
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

func (r *Tokenizer) SetReader(reader io.Reader) error {
	r.scanner.SetReader(reader)
	return nil
}

func (r *Tokenizer) setMaxTokenLength(length int) {
	r.maxTokenLength = length
}
