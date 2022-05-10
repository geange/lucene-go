package tokenattributes

// TypeAttribute A Token's lexical type. The Default value is "word".
type TypeAttribute interface {

	// Type Returns this Token's lexical type. Defaults to "word".
	//See Also: setType(String)
	Type() string

	// SetType Set the lexical type.
	// See Also: type()
	SetType(_type string)
}

const (
	DEFAULT_TYPE = "word"
)
