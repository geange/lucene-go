package standard

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenizer_IncrementToken(t *testing.T) {
	as := assert.New(t)

	text := "aaaa bbbb cccc dddd eeee ffff jjjjj"

	tokenizer := NewTokenizer()

	err := tokenizer.SetReader(bytes.NewReader([]byte(text)))
	assert.Nil(t, err)

	_, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	as.Equal("aaaa", string(tokenizer.AttributeSource().CharTerm().Buffer()))

	_, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	as.Equal("bbbb", string(tokenizer.AttributeSource().CharTerm().Buffer()))

	_, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	as.Equal("cccc", string(tokenizer.AttributeSource().CharTerm().Buffer()))

	_, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	as.Equal("dddd", string(tokenizer.AttributeSource().CharTerm().Buffer()))

	_, err = tokenizer.IncrementToken()
	assert.Nil(t, err)
	as.Equal("eeee", string(tokenizer.AttributeSource().CharTerm().Buffer()))
}
