package store

// ByteBuffersDataInput Base IndexInput implementation that uses an array of ByteBuffers to represent a file.
// Because Java's ByteBuffer uses an int to address the values, it's necessary to access a file greater Integer.MAX_VALUE in size using multiple byte buffers.
// For efficiency, this class requires that the buffers are a power-of-two (chunkSizePower).
type ByteBuffersDataInput interface {
	IndexInput
	//RandomAccessInput
}
