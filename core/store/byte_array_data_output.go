package store

var _ DataOutput = &ByteArrayDataOutput{}

// ByteArrayDataOutput DataOutput backed by a byte array. WARNING: This class omits most low-level checks, so be sure to test heavily with assertions enabled.
type ByteArrayDataOutput struct {
}
