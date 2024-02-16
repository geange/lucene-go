package packed

import "testing"

func TestBulkOperationPacked13_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 13, 100, NewBulkOperationPacked13())
}

func TestBulkOperationPacked13_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 13, 100, NewBulkOperationPacked13())
}
