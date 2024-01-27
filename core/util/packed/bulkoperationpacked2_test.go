package packed

import (
	"testing"
)

func TestBulkOperationPacked2_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 2, 100, NewBulkOperationPacked2())
}

func TestBulkOperationPacked2_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 2, 100, NewBulkOperationPacked2())
}
