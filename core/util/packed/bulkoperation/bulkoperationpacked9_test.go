package bulkoperation

import "testing"

func TestBulkOperationPacked9_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 9, 100, NewPacked9())
}

func TestBulkOperationPacked9_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 9, 100, NewPacked9())
}
