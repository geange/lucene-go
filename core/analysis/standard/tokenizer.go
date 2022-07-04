package standard

import (
	"github.com/geange/lucene-go/core/tokenattributes"
	"io"
)

type Tokenizer struct {
	source *tokenattributes.AttributeSourceV1

	scanner *TokenizerImpl

	skippedPositions int
	maxTokenLength   int
}

func NewTokenizer(reader io.Reader) *Tokenizer {
	tokenizer := &Tokenizer{
		source:           tokenattributes.NewAttributeSourceV1(),
		scanner:          &TokenizerImpl{},
		skippedPositions: 0,
		maxTokenLength:   0,
	}
	tokenizer.SetReader(reader)
	return tokenizer
}

func (r *Tokenizer) GetAttributeSource() *tokenattributes.AttributeSource {
	return nil
}

func (r *Tokenizer) AttributeSource() *tokenattributes.AttributeSourceV1 {
	return r.source
}

func (r *Tokenizer) IncrementToken() (bool, error) {
	r.source.Clear()
	r.skippedPositions = 0

	text, err := r.scanner.GetNextToken()
	if err != nil {
		return false, err
	}

	r.source.PackedTokenAttribute().Append(text)
	r.source.PackedTokenAttribute().SetOffset(r.scanner.Slow, r.scanner.Slow+len(text))
	return true, nil
}

func (r *Tokenizer) End() error {
	return nil
}

func (r *Tokenizer) Reset() error {
	return nil
}

func (r *Tokenizer) Close() error {
	return nil
}

func (r *Tokenizer) SetReader(reader io.Reader) error {
	r.scanner.SetReader(reader)
	return nil
}

func (r *Tokenizer) setMaxTokenLength(length int) {
	r.maxTokenLength = length
}
