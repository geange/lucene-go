package attribute

import (
	"errors"
)

type Factory interface {
	// CreateAttributeInstance Returns an AttributeImpl for the supplied Attribute interface class.
	CreateAttributeInstance(class string) (Attribute, error)
}

var (
	DEFAULT_ATTRIBUTE_FACTORY Factory = &DefaultAttributeFactory{}
)

type DefaultAttributeFactory struct {
}

func (d DefaultAttributeFactory) CreateAttributeInstance(class string) (Attribute, error) {
	switch class {
	case ClassBytesTerm:
		return newBytesTermAttr(), nil
	case ClassCharTerm:
		return NewPackedTokenAttr(), nil
	case ClassOffset:
		return NewPackedTokenAttr(), nil
	case ClassPositionIncrement:
		return NewPackedTokenAttr(), nil
	case ClassPositionLength:
		return NewPackedTokenAttr(), nil
	case ClassTermFrequency:
		return NewPackedTokenAttr(), nil
	case ClassTermToBytesRef:
		return NewPackedTokenAttr(), nil
	case ClassPayload:
		return newPayloadAttr(), nil
	default:
		return nil, errors.New("attribute not exist")
	}
}
