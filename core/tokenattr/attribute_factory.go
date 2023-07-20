package tokenattr

import (
	"errors"
)

type AttributeFactory interface {
	// CreateAttributeInstance Returns an AttributeImpl for the supplied Attribute interface class.
	// Throws:  UndeclaredThrowableException – A wrapper runtime exception thrown if the constructor of the
	// 		    attribute class throws a checked exception. Note that attributes should not throw or declare checked
	// 			exceptions; this may be verified and fail early in the future.
	CreateAttributeInstance(class string) (Attribute, error)
}

var (
	DEFAULT_ATTRIBUTE_FACTORY AttributeFactory = &DefaultAttributeFactory{}
)

type DefaultAttributeFactory struct {
}

func (d DefaultAttributeFactory) CreateAttributeInstance(class string) (Attribute, error) {
	switch class {
	case ClassBytesTerm:
		return NewBytesTermAttrBase(), nil
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
		return NewPayloadAttrBase(), nil
	default:
		return nil, errors.New("attribute not exist")
	}
}
