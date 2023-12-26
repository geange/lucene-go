package packed

import "testing"

func TestBulkOperationPacked3_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 3, 100, NewBulkOperationPacked3())
}

func TestBulkOperationPacked3_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 3, 100, NewBulkOperationPacked3())
}
