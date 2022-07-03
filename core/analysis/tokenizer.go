package analysis

import (
	"github.com/geange/lucene-go/core/util"
	"io"
)

type Tokenizer interface {
	TokenStream

	// SetReader Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in its tokenStream method)
	// will use this to re-use a previously created tokenizer.
	SetReader(reader io.Reader) error
}

func NewTokenizerImpl(source *util.AttributeSource) *TokenizerIMP {
	return &TokenizerIMP{
		source:       source,
		Input:        nil,
		inputPending: nil,
	}
}

type TokenizerIMP struct {
	source *util.AttributeSource

	sourceV1 *util.AttributeSourceV1

	// The text source for this Tokenizer.
	Input io.Reader

	// Pending reader: not actually assigned to input until reset()
	inputPending io.Reader
}

func (t *TokenizerIMP) AttributeSource() *util.AttributeSourceV1 {
	return t.sourceV1
}

func (t *TokenizerIMP) GetAttributeSource() *util.AttributeSource {
	return t.source
}

func (t *TokenizerIMP) IncrementToken() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TokenizerIMP) End() error {
	//TODO implement me
	panic("implement me")
}

func (t *TokenizerIMP) Reset() error {
	t.Input = t.inputPending
	t.inputPending = nil
	return nil
}

func (t *TokenizerIMP) Close() error {
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
func (t *TokenizerIMP) CorrectOffset(currentOff int) int {
	if charFilter, ok := t.Input.(CharFilter); ok {
		return charFilter.CorrectOffset(currentOff)
	}
	return currentOff
}

func (t *TokenizerIMP) SetReader(reader io.Reader) error {
	t.inputPending = reader
	return nil
}
