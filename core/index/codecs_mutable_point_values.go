package index

import "bytes"

type MutablePointValues interface {
	PointValues

	// GetValue Set packedValue with a reference to the packed bytes of the i-th value.
	GetValue(i int, packedValue *bytes.Buffer)

	// GetByteAt Get the k-th byte of the i-th value.
	GetByteAt(i, k int) byte

	// GetDocID Return the doc ID of the i-th value.
	GetDocID(i int) int

	// Swap the i-th and j-th values.
	Swap(i, j int)

	// Save the i-th value into the j-th position in temporary storage.
	Save(i, j int)

	// Restore values between i-th and j-th(excluding) in temporary storage into original storage.
	Restore(i, j int)
}
