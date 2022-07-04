package standard

import (
	"github.com/geange/lucene-go/core/util"
	"io"
)

type Tokenizer struct {
	scanner *TokenizerImpl
}

func NewTokenizer() *Tokenizer {
	panic("")
}

func (r *Tokenizer) GetAttributeSource() *util.AttributeSource {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) AttributeSource() *util.AttributeSourceV1 {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) IncrementToken() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) End() error {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) Reset() error {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) SetReader(reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (r *Tokenizer) setMaxTokenLength(length int) {
	panic("")
}
