package tokenattributes

var (
	_ CharTermAttribute = &CharTermAttributeImp{}
)

func NewCharTermAttributeImpl() *CharTermAttributeImp {
	return &CharTermAttributeImp{termBuffer: make([]rune, 0)}
}

type CharTermAttributeImp struct {
	termBuffer []rune
}

func (c *CharTermAttributeImp) AppendRune(r rune) {
	c.termBuffer = append(c.termBuffer, r)
}

func (c *CharTermAttributeImp) ResizeBuffer(newSize int) {
	if len(c.termBuffer) < newSize {
		c.termBuffer = make([]rune, newSize)
	}
}

func (c *CharTermAttributeImp) SetLength(length int) {
	if len(c.termBuffer) < length {
		c.termBuffer = c.termBuffer[:length]
		return
	}
	c.ResizeBuffer(length)
}

func (c *CharTermAttributeImp) Buffer() []rune {
	return c.termBuffer
}

func (c *CharTermAttributeImp) Append(s string) {
	c.termBuffer = append(c.termBuffer, []rune(s)...)
}

func (c *CharTermAttributeImp) SetEmpty() {
	c.termBuffer = c.termBuffer[:0]
}

func (c *CharTermAttributeImp) Interfaces() []string {
	return []string{
		"CharTerm",
		"TermToBytesRef",
	}
}

func (c *CharTermAttributeImp) Clear() error {
	c.termBuffer = c.termBuffer[:0]
	return nil
}

func (c *CharTermAttributeImp) End() error {
	return nil
}

func (c *CharTermAttributeImp) CopyTo(target AttributeImpl) error {
	impl, ok := target.(*CharTermAttributeImp)
	if ok {
		termBuffer := make([]rune, len(c.termBuffer))
		copy(termBuffer, c.termBuffer)
		impl.termBuffer = termBuffer
		return nil
	}
	return nil
}

func (c *CharTermAttributeImp) Clone() AttributeImpl {
	termBuffer := make([]rune, len(c.termBuffer))
	copy(termBuffer, c.termBuffer)
	return &CharTermAttributeImp{termBuffer: termBuffer}
}
