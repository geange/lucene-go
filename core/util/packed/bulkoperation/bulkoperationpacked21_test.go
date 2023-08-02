package bulkoperation

import "testing"

func TestBulkOperationPacked21_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 21, 100, NewPacked21())
}

func TestBulkOperationPacked21_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 21, 100, NewPacked21())
}
