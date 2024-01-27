package packed

import "testing"

func TestBulkOperationPacked7_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 7, 100, NewBulkOperationPacked7())
}

func TestBulkOperationPacked7_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 7, 100, NewBulkOperationPacked7())
}
