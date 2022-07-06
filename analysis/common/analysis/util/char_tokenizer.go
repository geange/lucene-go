package util

import (
	"bufio"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/tokenattributes"
	"io"
)

// CharTokenizer An abstract base class for simple, character-oriented tokenizers.
// The base class also provides factories to create instances of CharTokenizer using Java 8 lambdas or method
// references. It is possible to create an instance which behaves exactly like LetterTokenizer:
// Tokenizer tok = CharTokenizer.fromTokenCharPredicate(Character::isLetter);
type CharTokenizer interface {
	CharTokenizerExt

	analysis.Tokenizer
}

type CharTokenizerExt interface {
	IsTokenChar(r rune) bool
}

func NewCharTokenizerImpl(ext CharTokenizerExt, input io.Reader) *CharTokenizerImpl {
	tokenizer := analysis.NewTokenizerImpl(tokenattributes.NewAttributeSource())
	tokenizer.SetReader(input)
	tokenizer.Reset()

	attr := tokenattributes.NewPackedTokenAttributeImp()

	return &CharTokenizerImpl{
		ext:          ext,
		TokenizerImp: tokenizer,
		offset:       0,
		finalOffset:  0,
		maxTokenLen:  86400,
		termAtt:      attr,
		offsetAtt:    attr,
		reader:       bufio.NewReader(input),
	}
}

type CharTokenizerImpl struct {
	ext CharTokenizerExt

	*analysis.TokenizerImp

	offset      int //
	finalOffset int
	maxTokenLen int

	termAtt   tokenattributes.CharTermAttribute
	offsetAtt tokenattributes.OffsetAttribute

	reader *bufio.Reader
}

func (c *CharTokenizerImpl) IsTokenChar(r rune) bool {
	return c.ext.IsTokenChar(r)
}

// IncrementToken
// 每次读取 ioBuffer 的数据
func (c *CharTokenizerImpl) IncrementToken() (bool, error) {
	// clearAttributes();
	start, end := -1, -1

	for {
		r, size, err := c.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		c.offset += size

		if c.ext.IsTokenChar(r) {
			if start < 0 {
				start = c.finalOffset
			}
			c.finalOffset = c.offset
			end = c.finalOffset
			c.termAtt.AppendRune(r)
		} else if start >= 0 {
			break
		}
	}

	if err := c.offsetAtt.SetOffset(start, end); err != nil {
		return false, err
	}
	return true, nil
}

func (c *CharTokenizerImpl) End() error {
	return c.offsetAtt.SetOffset(c.finalOffset, c.finalOffset)
}

func (c *CharTokenizerImpl) Reset() error {
	err := c.TokenizerImp.Reset()
	if err != nil {
		return err
	}

	c.offset = 0
	c.finalOffset = 0

	return nil
}

func (c *CharTokenizerImpl) Close() error {
	//TODO implement me
	panic("implement me")
}

func (c *CharTokenizerImpl) SetReader(reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}
