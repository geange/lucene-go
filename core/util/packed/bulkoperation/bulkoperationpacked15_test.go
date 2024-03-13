package bulkoperation

import "testing"

func TestBulkOperationPacked15_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 15, 100, NewPacked15())
}

func TestBulkOperationPacked15_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 15, 100, NewPacked15())
}
