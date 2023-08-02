package fst

import "github.com/pkg/errors"

var (
	ErrByteStoreBasic = errors.New("bytestore basic error")
	ErrItemNotFound   = errors.Wrap(ErrByteStoreBasic, "item not found")
)
