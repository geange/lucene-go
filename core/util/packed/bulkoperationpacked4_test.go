package packed

import "testing"

func TestBulkOperationPacked4_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 4, 100, NewBulkOperationPacked4())
}

func TestBulkOperationPacked4_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 4, 100, NewBulkOperationPacked4())
}
