package packed

import "testing"

func TestBulkOperationPacked_DecodeUint64(t *testing.T) {
	t.Run("bitsPerValue = 25", func(t *testing.T) {
		testDecodeUint64(t, 64, 25, 200, NewBulkOperationPacked(25))
	})

	t.Run("bitsPerValue = 26", func(t *testing.T) {
		testDecodeUint64(t, 64, 26, 200, NewBulkOperationPacked(26))
	})

	t.Run("bitsPerValue = 27", func(t *testing.T) {
		testDecodeUint64(t, 64, 27, 200, NewBulkOperationPacked(27))
	})

	t.Run("bitsPerValue = 28", func(t *testing.T) {
		testDecodeUint64(t, 64, 28, 200, NewBulkOperationPacked(28))
	})

	t.Run("bitsPerValue = 29", func(t *testing.T) {
		testDecodeUint64(t, 64, 29, 200, NewBulkOperationPacked(29))
	})

	t.Run("bitsPerValue = 30", func(t *testing.T) {
		testDecodeUint64(t, 64, 30, 200, NewBulkOperationPacked(30))
	})
}

func TestBulkOperationPacked_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 25, 100, NewBulkOperationPacked(25))
}
