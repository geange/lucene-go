package store

var _ DataOutput = &ByteBuffersDataOutput{}

// ByteBuffersDataOutput A DataOutput storing data in a list of ByteBuffers.
type ByteBuffersDataOutput struct {
}
