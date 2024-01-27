package packed

import "testing"

func TestBulkOperationPacked19_DecodeUint64(t *testing.T) {
	testDecodeUint64(t, 64, 19, 100, NewBulkOperationPacked19())
}

func TestBulkOperationPacked19_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 19, 100, NewBulkOperationPacked19())
}
