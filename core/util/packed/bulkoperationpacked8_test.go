package packed

import "testing"

func TestBulkOperationPacked8_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 8, 100, NewBulkOperationPacked8())
}

func TestBulkOperationPacked8_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 8, 100, NewBulkOperationPacked8())
}
