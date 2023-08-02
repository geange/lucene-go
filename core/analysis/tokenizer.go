package analysis

import (
	"io"

	"github.com/geange/lucene-go/core/util/attribute"
)

type Tokenizer interface {
	TokenStream

	// SetReader Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in its tokenStream method)
	// will use this to re-use a previously created tokenizer.
	SetReader(reader io.Reader) error
}

func NewBaseTokenizer() *BaseTokenizer {
	return &BaseTokenizer{
		source:       attribute.NewSource(),
		input:        nil,
		inputPending: nil,
	}
}

type BaseTokenizer struct {
	source       *attribute.Source
	input        io.Reader // The text source for this Tokenizer.
	inputPending io.Reader // Pending reader: not actually assigned to input until reset()
}

func (t *BaseTokenizer) AttributeSource() *attribute.Source {
	return t.source
}

func (t *BaseTokenizer) End() error {
	return nil
}

func (t *BaseTokenizer) Reset() error {
	t.input = t.inputPending
	t.inputPending = nil
	return nil
}

func (t *BaseTokenizer) Close() error {
	t.input = nil
	t.inputPending = nil
	return nil
}

// CorrectOffset Return the corrected offset. If input is a CharFilter subclass this method calls
// CharFilter.correctOffset, else returns currentOff.
// Params: currentOff – offset as seen in the output
// Returns: corrected offset based on the input
// See Also: CharFilter.correctOffset
func (t *BaseTokenizer) CorrectOffset(currentOff int) int {
	if charFilter, ok := t.input.(CharFilter); ok {
		return charFilter.CorrectOffset(currentOff)
	}
	return currentOff
}

func (t *BaseTokenizer) SetReader(reader io.Reader) error {
	t.inputPending = reader
	return nil
}
