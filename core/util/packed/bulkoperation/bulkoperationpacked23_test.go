package bulkoperation

import "testing"

func TestBulkOperationPacked23_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 23, 100, NewPacked23())
}

func TestBulkOperationPacked23_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 23, 100, NewPacked23())
}
