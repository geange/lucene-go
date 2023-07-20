package util

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharTokenizerImpl_IncrementToken(t *testing.T) {
	text := "a b ccc dddd"

	tokenizer := NewCharTokenizerImpl(&ext{}, bytes.NewReader([]byte(text)))

	ok, err := tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, []rune("a"), tokenizer.termAtt.Buffer())
	tokenizer.termAtt.SetEmpty()

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, []rune("b"), tokenizer.termAtt.Buffer())
	tokenizer.termAtt.SetEmpty()

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, []rune("ccc"), tokenizer.termAtt.Buffer())
	tokenizer.termAtt.SetEmpty()

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	assert.Equal(t, []rune("dddd"), tokenizer.termAtt.Buffer())
	tokenizer.termAtt.SetEmpty()
}

type ext struct {
}

func (e *ext) IsTokenChar(r rune) bool {
	return r != ' '
}
