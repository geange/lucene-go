package standard

import (
	"bytes"
	"testing"
)

func TestTokenizer_IncrementToken(t *testing.T) {
	text := "aaaa bbbb cccc dddd eeee ffff jjjjj"

	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))

	tokenizer.IncrementToken()
	tokenizer.IncrementToken()
	tokenizer.IncrementToken()
	tokenizer.IncrementToken()
	tokenizer.IncrementToken()
}
