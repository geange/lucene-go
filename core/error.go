package core

import "errors"

var (
	FrrFieldNotFound        = errors.New("field not found")
	ErrFieldValueTypeNotFit = errors.New("field value types not fit")
)
