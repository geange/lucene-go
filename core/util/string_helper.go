package util

import "errors"

func BytesDifference(priorTerm, currentTerm []byte) (int, error) {
	mismatch := Mismatch(priorTerm, currentTerm)
	if mismatch < 0 {
		return -1, errors.New("terms out of order")
	}
	return mismatch, nil
}
