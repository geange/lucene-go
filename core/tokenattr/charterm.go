package tokenattr

var (
	_ CharTermAttribute = &CharTermAttrBase{}
)

func NewCharTermAttr() *CharTermAttrBase {
	return &CharTermAttrBase{termBuffer: make([]rune, 0)}
}

type CharTermAttrBase struct {
	termBuffer []rune
}

func (c *CharTermAttrBase) AppendRune(r rune) {
	c.termBuffer = append(c.termBuffer, r)
}

func (c *CharTermAttrBase) ResizeBuffer(newSize int) {
	if len(c.termBuffer) < newSize {
		c.termBuffer = make([]rune, newSize)
	}
}

func (c *CharTermAttrBase) SetLength(length int) {
	if len(c.termBuffer) < length {
		c.termBuffer = c.termBuffer[:length]
		return
	}
	c.ResizeBuffer(length)
}

func (c *CharTermAttrBase) Buffer() []rune {
	return c.termBuffer
}

func (c *CharTermAttrBase) Append(s string) {
	c.termBuffer = append(c.termBuffer, []rune(s)...)
}

func (c *CharTermAttrBase) SetEmpty() {
	c.termBuffer = c.termBuffer[:0]
}

func (c *CharTermAttrBase) Interfaces() []string {
	return []string{
		"CharTerm",
		"TermToBytesRef",
	}
}

func (c *CharTermAttrBase) Clear() error {
	c.termBuffer = c.termBuffer[:0]
	return nil
}

func (c *CharTermAttrBase) End() error {
	return nil
}

func (c *CharTermAttrBase) CopyTo(target Attribute) error {
	impl, ok := target.(*CharTermAttrBase)
	if ok {
		termBuffer := make([]rune, len(c.termBuffer))
		copy(termBuffer, c.termBuffer)
		impl.termBuffer = termBuffer
		return nil
	}
	return nil
}

func (c *CharTermAttrBase) Clone() Attribute {
	termBuffer := make([]rune, len(c.termBuffer))
	copy(termBuffer, c.termBuffer)
	return &CharTermAttrBase{termBuffer: termBuffer}
}
