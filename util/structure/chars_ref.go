package structure

type CharsRef struct {
	Chars []rune
}

func NewCharsRef(chars []rune) *CharsRef {
	return &CharsRef{Chars: chars}
}

func (c *CharsRef) Len() int {
	return len(c.Chars)
}
