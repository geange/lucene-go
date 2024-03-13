package bulkoperation

import "testing"

func TestBulkOperationPacked22_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 22, 100, NewPacked22())
}

func TestBulkOperationPacked22_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 22, 100, NewPacked22())
}
