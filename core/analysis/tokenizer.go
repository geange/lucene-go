package analysis

import (
	"github.com/geange/lucene-go/core/tokenattributes"
	"io"
)

type Tokenizer interface {
	TokenStream

	// SetReader Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in its tokenStream method)
	// will use this to re-use a previously created tokenizer.
	SetReader(reader io.Reader) error
}

func NewTokenizerImpl(source *tokenattributes.AttributeSourceV2) *TokenizerImp {
	return &TokenizerImp{
		source:       source,
		Input:        nil,
		inputPending: nil,
	}
}

type TokenizerImp struct {
	source *tokenattributes.AttributeSourceV2

	sourceV1 *tokenattributes.AttributeSource

	// The text source for this Tokenizer.
	Input io.Reader

	// Pending reader: not actually assigned to input until reset()
	inputPending io.Reader
}

func (t *TokenizerImp) AttributeSource() *tokenattributes.AttributeSource {
	return t.sourceV1
}

func (t *TokenizerImp) GetAttributeSource() *tokenattributes.AttributeSourceV2 {
	return t.source
}

func (t *TokenizerImp) IncrementToken() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TokenizerImp) End() error {
	//TODO implement me
	panic("implement me")
}

func (t *TokenizerImp) Reset() error {
	t.Input = t.inputPending
	t.inputPending = nil
	return nil
}

func (t *TokenizerImp) Close() error {
	//err := t.Input.Close()
	//if err != nil {
	//	return err
	//}

	t.Input = nil
	t.inputPending = nil
	return nil
}

// CorrectOffset Return the corrected offset. If input is a CharFilter subclass this method calls
// CharFilter.correctOffset, else returns currentOff.
// Params: currentOff – offset as seen in the output
// Returns: corrected offset based on the input
// See Also: CharFilter.correctOffset
func (t *TokenizerImp) CorrectOffset(currentOff int) int {
	if charFilter, ok := t.Input.(CharFilter); ok {
		return charFilter.CorrectOffset(currentOff)
	}
	return currentOff
}

func (t *TokenizerImp) SetReader(reader io.Reader) error {
	t.inputPending = reader
	return nil
}
