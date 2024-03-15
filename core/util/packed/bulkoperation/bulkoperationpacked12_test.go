package bulkoperation

import "testing"

func TestBulkOperationPacked12_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 12, 100, NewPacked12())
}

func TestBulkOperationPacked12_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 12, 100, NewPacked12())
}
