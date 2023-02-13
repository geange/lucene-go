package numeric

import (
	"errors"
)

// Subtract Result = a - b, where a >= b, else IllegalArgumentException is thrown.
func Subtract(bytesPerDim, dim int, a, b, result []byte) error {
	start := dim * bytesPerDim
	end := start + bytesPerDim
	borrow := 0
	for i := end - 1; i >= start; i-- {
		diff := int(a[i]) - int(b[i]&0xff) - borrow
		if diff < 0 {
			diff += 256
			borrow = 1
		} else {
			borrow = 0
		}
		result[i-start] = byte(diff)
	}
	if borrow != 0 {
		return errors.New("a < b")
	}
	return nil
}
