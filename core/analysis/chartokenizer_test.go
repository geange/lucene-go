package analysis

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharTokenizerImpl_IncrementToken(t *testing.T) {
	text := "a b ccc dddd"

	tokenizer := NewCharTokenizerImpl(&ext{}, bytes.NewReader([]byte(text)))

	ok, err := tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, "a", tokenizer.termAtt.GetString())
	tokenizer.termAtt.Reset()

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, "b", tokenizer.termAtt.GetString())
	tokenizer.termAtt.Reset()

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, "ccc", tokenizer.termAtt.GetString())
	tokenizer.termAtt.Reset()

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, "dddd", tokenizer.termAtt.GetString())
	tokenizer.termAtt.Reset()
}

type ext struct {
}

func (e *ext) IsTokenChar(r rune) bool {
	return r != ' '
}
