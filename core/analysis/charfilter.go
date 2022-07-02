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

func NewCharFilterImpl(ext CharFilterExt, input io.ReadCloser) *CharFilterImpl {
	return &CharFilterImpl{
		ext:   ext,
		input: input,
	}
}

type CharFilterImpl struct {
	ext   CharFilterExt
	input io.ReadCloser
}

func (c *CharFilterImpl) Correct(currentOff int) int {
	return c.ext.Correct(currentOff)
}

func (c *CharFilterImpl) Close() error {
	return c.input.Close()
}

func (c *CharFilterImpl) Read(p []byte) (n int, err error) {
	return c.input.Read(p)
}

func (c *CharFilterImpl) CorrectOffset(currentOff int) int {
	corrected := c.ext.Correct(currentOff)
	if charFilter, ok := c.input.(CharFilter); ok {
		return charFilter.CorrectOffset(corrected)
	}
	return corrected
}
