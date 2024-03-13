package bulkoperation

import "testing"

func TestBulkOperationPacked11_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 11, 100, NewPacked11())
}

func TestBulkOperationPacked11_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 11, 100, NewPacked11())
}
