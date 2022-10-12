package fst

import "errors"

var (
	ErrOutOfArrayRange      = errors.New("out of array range")
	ErrIllegalArgument      = errors.New("illegal argument")
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrIllegalState         = errors.New("illegal state")
)
