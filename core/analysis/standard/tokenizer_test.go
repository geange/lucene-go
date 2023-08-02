package standard

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizer_IncrementToken(t *testing.T) {
	as := assert.New(t)

	text := "aaaa bbbb cccc dddd eeee"

	tokenizer := NewTokenizer()

	err := tokenizer.SetReader(bytes.NewReader([]byte(text)))
	assert.Nil(t, err)

	ok, err := tokenizer.IncrementToken()
	assert.Nil(t, err)
	assert.True(t, ok)
	as.Equal("aaaa", string(tokenizer.AttributeSource().CharTerm().GetString()))

	ok, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	assert.True(t, ok)
	as.Equal("bbbb", string(tokenizer.AttributeSource().CharTerm().GetString()))

	ok, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	assert.True(t, ok)
	as.Equal("cccc", string(tokenizer.AttributeSource().CharTerm().GetString()))

	ok, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	assert.True(t, ok)
	as.Equal("dddd", string(tokenizer.AttributeSource().CharTerm().GetString()))

	ok, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	assert.True(t, ok)
	as.Equal("eeee", string(tokenizer.AttributeSource().CharTerm().GetString()))

	ok, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	assert.False(t, ok)
}
