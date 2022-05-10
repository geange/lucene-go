package tokenattributes

type PositionLengthAttribute interface {
	// SetPositionLength Set the position length of this Token.
	// The default value is one.
	// Params: positionLength – how many positions this token spans.
	// Throws: IllegalArgumentException – if positionLength is zero or negative.
	// See Also: getPositionLength()
	SetPositionLength(positionLength int) error

	// GetPositionLength Returns the position length of this Token.
	// See Also: setPositionLength
	GetPositionLength() int
}
