package packed

import "testing"

func TestBulkOperationPacked16_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 16, 100, NewBulkOperationPacked16())
}

func TestBulkOperationPacked16_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 16, 100, NewBulkOperationPacked16())
}
