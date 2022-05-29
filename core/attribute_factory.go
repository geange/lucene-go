package core

import (
	"errors"
)

type AttributeFactory interface {
	// CreateAttributeInstance Returns an AttributeImpl for the supplied Attribute interface class.
	// Throws:  UndeclaredThrowableException – A wrapper runtime exception thrown if the constructor of the
	// 		    attribute class throws a checked exception. Note that attributes should not throw or declare checked
	// 			exceptions; this may be verified and fail early in the future.
	CreateAttributeInstance(class string) (AttributeImpl, error)
}

var (
	DEFAULT_ATTRIBUTE_FACTORY AttributeFactory = &DefaultAttributeFactory{}
)

type DefaultAttributeFactory struct {
}

func (d DefaultAttributeFactory) CreateAttributeInstance(class string) (AttributeImpl, error) {
	switch class {
	case ClassBytesTerm:
		return NewBytesTermAttributeImpl(), nil
	case ClassCharTerm:
		return NewPackedTokenAttributeImpl(), nil
	case ClassOffset:
		return NewPackedTokenAttributeImpl(), nil
	case ClassPositionIncrement:
		return NewPackedTokenAttributeImpl(), nil
	case ClassPositionLength:
		return NewPackedTokenAttributeImpl(), nil
	case ClassTermFrequency:
		return NewPackedTokenAttributeImpl(), nil
	case ClassTermToBytesRef:
		return NewPackedTokenAttributeImpl(), nil
	case ClassPayload:
		return NewPayloadAttributeImpl(), nil
	default:
		return nil, errors.New("attribute not exist")
	}
}
