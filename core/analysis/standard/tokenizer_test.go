package standard

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenizer_IncrementToken(t *testing.T) {
	assert := assert.New(t)

	text := "aaaa bbbb cccc dddd eeee ffff jjjjj"

	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))

	tokenizer.IncrementToken()
	assert.Equal("aaaa", string(tokenizer.AttributeSource().PackedTokenAttribute().Buffer()))
	tokenizer.IncrementToken()
	assert.Equal("bbbb", string(tokenizer.AttributeSource().PackedTokenAttribute().Buffer()))
	tokenizer.IncrementToken()
	assert.Equal("cccc", string(tokenizer.AttributeSource().PackedTokenAttribute().Buffer()))
	tokenizer.IncrementToken()
	assert.Equal("dddd", string(tokenizer.AttributeSource().PackedTokenAttribute().Buffer()))
	tokenizer.IncrementToken()
	assert.Equal("eeee", string(tokenizer.AttributeSource().PackedTokenAttribute().Buffer()))
}
