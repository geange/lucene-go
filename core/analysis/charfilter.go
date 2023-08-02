package analysis

import "io"

type CharFilter interface {
	CharFilterExt

	io.ReadCloser

	// CorrectOffset Chains the corrected offset through the input CharFilter(s).
	CorrectOffset(currentOff int) int
}

type CharFilterExt interface {
	// Correct Subclasses override to correct the current offset.
	// Params: currentOff – current offset
	// Returns: corrected offset
	Correct(currentOff int) int
}

func NewBaseCharFilter(ext CharFilterExt, input io.ReadCloser) *BaseCharFilter {
	return &BaseCharFilter{
		ext:   ext,
		input: input,
	}
}

type BaseCharFilter struct {
	ext   CharFilterExt
	input io.ReadCloser
}

func (c *BaseCharFilter) Correct(currentOff int) int {
	return c.ext.Correct(currentOff)
}

func (c *BaseCharFilter) Close() error {
	return c.input.Close()
}

func (c *BaseCharFilter) Read(p []byte) (n int, err error) {
	return c.input.Read(p)
}

func (c *BaseCharFilter) CorrectOffset(currentOff int) int {
	corrected := c.ext.Correct(currentOff)
	if charFilter, ok := c.input.(CharFilter); ok {
		return charFilter.CorrectOffset(corrected)
	}
	return corrected
}
