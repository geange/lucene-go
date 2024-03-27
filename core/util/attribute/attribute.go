package attribute

const (
	ClassBytesTerm         = "BytesTerm"
	ClassCharTerm          = "CharTerm"
	ClassOffset            = "Offset"
	ClassPositionIncrement = "PositionIncrement"
	ClassPayload           = "Payload"
	ClassPositionLength    = "PositionLength"
	ClassTermFrequency     = "TermFrequency"
	ClassTermToBytesRef    = "TermToBytesRef"
	ClassType              = "Type"
)

// Attribute
// Base class for Attributes that can be added to a AttributeSourceV2.
// Attributes are used to add data in a dynamic, yet types-safe way to a source of usually streamed objects,
type Attribute interface {
	Interfaces() []string
	Reset() error
	CopyTo(target Attribute) error
	Clone() Attribute
}

const (
	DEFAULT_TYPE = "word"
)
