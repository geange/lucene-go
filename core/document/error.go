package document

import "errors"

var (
	FrrFieldNotFound        = errors.New("field not found")
	ErrFieldValueTypeNotFit = errors.New("field value type not fit")
)