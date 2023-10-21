package util

import (
	"errors"
	"github.com/geange/lucene-go/core/util/array"
)

func BytesDifference(priorTerm, currentTerm []byte) (int, error) {
	mismatch := array.Mismatch(priorTerm, currentTerm)
	if mismatch < 0 {
		return -1, errors.New("terms out of order")
	}
	return mismatch, nil
}
