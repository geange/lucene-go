package util

import (
	"bufio"
	"io"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/tokenattr"
)

// CharTokenizer An abstract base class for simple, character-oriented tokenizers.
// The base class also provides factories to create instances of CharTokenizer using Java 8 lambdas or method
// references. It is possible to create an instance which behaves exactly like LetterTokenizer:
// Tokenizer tok = CharTokenizer.fromTokenCharPredicate(Character::isLetter);
type CharTokenizer interface {
	CharTokenizerInner

	analysis.Tokenizer
}

type CharTokenizerInner interface {
	IsTokenChar(r rune) bool
}

func NewCharTokenizerImpl(ext CharTokenizerInner, input io.Reader) *CharTokenizerBase {
	tokenizer := analysis.NewTokenizer()
	tokenizer.SetReader(input)
	tokenizer.Reset()

	attr := tokenattr.NewPackedTokenAttr()

	return &CharTokenizerBase{
		inner:         ext,
		TokenizerBase: tokenizer,
		offset:        0,
		finalOffset:   0,
		maxTokenLen:   86400,
		termAtt:       attr,
		offsetAtt:     attr,
		reader:        bufio.NewReader(input),
	}
}

type CharTokenizerBase struct {
	inner CharTokenizerInner

	*analysis.TokenizerBase

	offset      int //
	finalOffset int
	maxTokenLen int

	termAtt   tokenattr.CharTermAttr
	offsetAtt tokenattr.OffsetAttr

	reader *bufio.Reader

	isEOF bool
}

func (c *CharTokenizerBase) IsTokenChar(r rune) bool {
	return c.inner.IsTokenChar(r)
}

// IncrementToken
// 每次读取 ioBuffer 的数据
func (c *CharTokenizerBase) IncrementToken() (bool, error) {
	if c.isEOF {
		return false, io.EOF
	}

	// clearAttributes();
	start, end := -1, -1

	for {
		r, size, err := c.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				c.isEOF = true
				return true, nil
			}
			return false, err
		}
		c.offset += size

		if c.inner.IsTokenChar(r) {
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

func (c *CharTokenizerBase) End() error {
	return c.offsetAtt.SetOffset(c.finalOffset, c.finalOffset)
}

func (c *CharTokenizerBase) Reset() error {
	err := c.TokenizerBase.Reset()
	if err != nil {
		return err
	}

	c.offset = 0
	c.finalOffset = 0

	return nil
}

func (c *CharTokenizerBase) Close() error {
	//TODO implement me
	panic("implement me")
}

func (c *CharTokenizerBase) SetReader(reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}
