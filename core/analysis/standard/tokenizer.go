package standard

import (
	"github.com/geange/lucene-go/core/util"
	"io"
)

type Tokenizer struct {
	scanner *TokenizerImpl
}

func (t *Tokenizer) GetAttributeSource() *util.AttributeSource {
	//TODO implement me
	panic("implement me")
}

func (t *Tokenizer) AttributeSource() *util.AttributeSourceV1 {
	//TODO implement me
	panic("implement me")
}

func (t *Tokenizer) IncrementToken() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *Tokenizer) End() error {
	//TODO implement me
	panic("implement me")
}

func (t *Tokenizer) Reset() error {
	//TODO implement me
	panic("implement me")
}

func (t *Tokenizer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *Tokenizer) SetReader(reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}
