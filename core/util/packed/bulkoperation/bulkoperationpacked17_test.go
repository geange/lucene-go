package bulkoperation

import "testing"

func TestBulkOperationPacked17_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 17, 100, NewPacked17())
}

func TestBulkOperationPacked17_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 17, 100, NewPacked17())
}
