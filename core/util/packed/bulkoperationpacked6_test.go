package packed

import "testing"

func TestBulkOperationPacked6_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 6, 100, NewBulkOperationPacked6())
}

func TestBulkOperationPacked6_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 6, 100, NewBulkOperationPacked6())
}
