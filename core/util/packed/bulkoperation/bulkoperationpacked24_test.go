package bulkoperation

import "testing"

func TestBulkOperationPacked24_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 24, 100, NewPacked24())
}

func TestBulkOperationPacked24_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 24, 100, NewPacked24())
}
