package util

import (
	"errors"
	"github.com/geange/lucene-go/core/analysis/tokenattributes"
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
	case tokenattributes.ClassBytesTerm:
		return tokenattributes.NewBytesTermAttributeImpl(), nil
	case tokenattributes.ClassCharTerm:
		return tokenattributes.NewPackedTokenAttributeImpl(), nil
	case tokenattributes.ClassOffset:
		return tokenattributes.NewPackedTokenAttributeImpl(), nil
	case tokenattributes.ClassPositionIncrement:
		return tokenattributes.NewPackedTokenAttributeImpl(), nil
	case tokenattributes.ClassPositionLength:
		return tokenattributes.NewPackedTokenAttributeImpl(), nil
	case tokenattributes.ClassTermFrequency:
		return tokenattributes.NewPackedTokenAttributeImpl(), nil
	case tokenattributes.ClassTermToBytesRef:
		return tokenattributes.NewPackedTokenAttributeImpl(), nil
	default:
		return nil, errors.New("attribute not exist")
	}
}
