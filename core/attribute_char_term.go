package core

var (
	_ CharTermAttribute = &CharTermAttributeImpl{}
)

func NewCharTermAttributeImpl() *CharTermAttributeImpl {
	return &CharTermAttributeImpl{termBuffer: make([]rune, 0)}
}

type CharTermAttributeImpl struct {
	termBuffer []rune
}

func (c *CharTermAttributeImpl) AppendRune(r rune) {
	c.termBuffer = append(c.termBuffer, r)
}

func (c *CharTermAttributeImpl) ResizeBuffer(newSize int) {
	if len(c.termBuffer) < newSize {
		c.termBuffer = make([]rune, newSize)
	}
}

func (c *CharTermAttributeImpl) SetLength(length int) {
	if len(c.termBuffer) < length {
		c.termBuffer = c.termBuffer[:length]
		return
	}
	c.ResizeBuffer(length)
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
	c.termBuffer = nil
	return nil
}

func (c *CharTermAttributeImpl) End() error {
	return nil
}

func (c *CharTermAttributeImpl) CopyTo(target AttributeImpl) error {
	impl, ok := target.(*CharTermAttributeImpl)
	if ok {
		termBuffer := make([]rune, len(c.termBuffer))
		copy(termBuffer, c.termBuffer)
		impl.termBuffer = termBuffer
		return nil
	}
	return nil
}

func (c *CharTermAttributeImpl) Clone() AttributeImpl {
	termBuffer := make([]rune, len(c.termBuffer))
	copy(termBuffer, c.termBuffer)
	return &CharTermAttributeImpl{termBuffer: termBuffer}
}
