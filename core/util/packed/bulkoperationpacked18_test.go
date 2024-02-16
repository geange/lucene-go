package packed

import "testing"

func TestBulkOperationPacked18_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 18, 100, NewBulkOperationPacked18())
}

func TestBulkOperationPacked18_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 18, 100, NewBulkOperationPacked18())
}
