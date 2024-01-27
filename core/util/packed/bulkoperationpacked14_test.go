package packed

import "testing"

func TestBulkOperationPacked14_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 14, 100, NewBulkOperationPacked14())
}

func TestBulkOperationPacked14_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 14, 100, NewBulkOperationPacked14())
}
