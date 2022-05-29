package util

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharTokenizerImpl_IncrementToken(t *testing.T) {
	text := "a b ccc ddd"

	tokenizer := NewCharTokenizerImpl(&ext{}, bytes.NewReader([]byte(text)))

	ok, err := tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	fmt.Println(string(tokenizer.termAtt.Buffer()))

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	fmt.Println(string(tokenizer.termAtt.Buffer()))

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	fmt.Println(string(tokenizer.termAtt.Buffer()))

	ok, err = tokenizer.IncrementToken()
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)
	fmt.Println(string(tokenizer.termAtt.Buffer()))

}

type ext struct {
}

func (e *ext) IsTokenChar(r rune) bool {
	return r != ' '
}
