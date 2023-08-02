package document

import "errors"

var (
	ErrFieldNotFound        = errors.New("field not found")
	ErrFieldValueTypeNotFit = errors.New("field value types not fit")
	ErrIllegalOperation     = errors.New("illegal operation")
	ErrFieldValueType       = errors.New("field value type not fit")
)
