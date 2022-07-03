package tokenattributes

import (
	"github.com/geange/lucene-go/core/util"
)

var (
	_ CharTermAttribute = &CharTermAttributeIMP{}
)

func NewCharTermAttributeImpl() *CharTermAttributeIMP {
	return &CharTermAttributeIMP{termBuffer: make([]rune, 0)}
}

type CharTermAttributeIMP struct {
	termBuffer []rune
}

func (c *CharTermAttributeIMP) AppendRune(r rune) {
	c.termBuffer = append(c.termBuffer, r)
}

func (c *CharTermAttributeIMP) ResizeBuffer(newSize int) {
	if len(c.termBuffer) < newSize {
		c.termBuffer = make([]rune, newSize)
	}
}

func (c *CharTermAttributeIMP) SetLength(length int) {
	if len(c.termBuffer) < length {
		c.termBuffer = c.termBuffer[:length]
		return
	}
	c.ResizeBuffer(length)
}

func (c *CharTermAttributeIMP) Buffer() []rune {
	return c.termBuffer
}

func (c *CharTermAttributeIMP) Append(s string) {
	c.termBuffer = append(c.termBuffer, []rune(s)...)
}

func (c *CharTermAttributeIMP) SetEmpty() {
	c.termBuffer = c.termBuffer[:0]
}

func (c *CharTermAttributeIMP) Interfaces() []string {
	return []string{
		"CharTerm",
		"TermToBytesRef",
	}
}

func (c *CharTermAttributeIMP) Clear() error {
	c.termBuffer = nil
	return nil
}

func (c *CharTermAttributeIMP) End() error {
	return nil
}

func (c *CharTermAttributeIMP) CopyTo(target util.AttributeImpl) error {
	impl, ok := target.(*CharTermAttributeIMP)
	if ok {
		termBuffer := make([]rune, len(c.termBuffer))
		copy(termBuffer, c.termBuffer)
		impl.termBuffer = termBuffer
		return nil
	}
	return nil
}

func (c *CharTermAttributeIMP) Clone() util.AttributeImpl {
	termBuffer := make([]rune, len(c.termBuffer))
	copy(termBuffer, c.termBuffer)
	return &CharTermAttributeIMP{termBuffer: termBuffer}
}
