package analysis

import (
	"github.com/geange/lucene-go/core/tokenattr"
	"io"
)

type Tokenizer interface {
	TokenStream

	// SetReader Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in its tokenStream method)
	// will use this to re-use a previously created tokenizer.
	SetReader(reader io.Reader) error
}

func NewTokenizer() *TokenizerBase {
	return &TokenizerBase{
		source:       tokenattr.NewAttributeSource(),
		Input:        nil,
		inputPending: nil,
	}
}

type TokenizerBase struct {
	source *tokenattr.AttributeSource

	// The text source for this Tokenizer.
	Input io.Reader

	// Pending reader: not actually assigned to input until reset()
	inputPending io.Reader
}

func (t *TokenizerBase) AttributeSource() *tokenattr.AttributeSource {
	return t.source
}

func (t *TokenizerBase) End() error {
	return nil
}

func (t *TokenizerBase) Reset() error {
	t.Input = t.inputPending
	t.inputPending = nil
	return nil
}

func (t *TokenizerBase) Close() error {
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
func (t *TokenizerBase) CorrectOffset(currentOff int) int {
	if charFilter, ok := t.Input.(CharFilter); ok {
		return charFilter.CorrectOffset(currentOff)
	}
	return currentOff
}

func (t *TokenizerBase) SetReader(reader io.Reader) error {
	t.inputPending = reader
	return nil
}
