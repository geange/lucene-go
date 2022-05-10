package tokenattributes

import "github.com/geange/lucene-go/core/util"

var (
	_ CharTermAttribute = &CharTermAttributeImpl{}
)

type CharTermAttributeImpl struct {
	termBuffer []rune
}

func (c *CharTermAttributeImpl) Buffer() []rune {
	return c.termBuffer
}

func (c *CharTermAttributeImpl) Append(s string) {
	c.termBuffer = append(c.termBuffer, []rune(s)...)
}

func (c *CharTermAttributeImpl) SetEmpty() {
	c.termBuffer = c.termBuffer[:0]
}

func (c *CharTermAttributeImpl) Interfaces() []string {
	return []string{
		"CharTerm",
		"TermToBytesRef",
	}
}

func (c *CharTermAttributeImpl) Clear() error {
	//TODO implement me
	panic("implement me")
}

func (c *CharTermAttributeImpl) End() error {
	//TODO implement me
	panic("implement me")
}

func (c *CharTermAttributeImpl) CopyTo(target util.AttributeImpl) error {
	//TODO implement me
	panic("implement me")
}

func (c *CharTermAttributeImpl) Clone() util.AttributeImpl {
	//TODO implement me
	panic("implement me")
}
