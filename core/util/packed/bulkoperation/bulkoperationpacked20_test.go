package bulkoperation

import "testing"

func TestBulkOperationPacked20_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 20, 100, NewPacked20())
}

func TestBulkOperationPacked20_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 20, 100, NewPacked20())
}
