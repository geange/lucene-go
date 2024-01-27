package packed

import "testing"

func TestBulkOperationPacked5_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 5, 100, NewBulkOperationPacked5())
}

func TestBulkOperationPacked5_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 5, 100, NewBulkOperationPacked5())
}
