package packed

import "testing"

func TestBulkOperationPacked10_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 10, 100, NewBulkOperationPacked10())
}

func TestBulkOperationPacked10_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 10, 100, NewBulkOperationPacked10())
}
