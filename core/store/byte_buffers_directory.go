package store

var _ BaseDirectory = &ByteBuffersDirectory{}

// ByteBuffersDirectory A ByteBuffer-based Directory implementation that can be used to store index files on the heap.
// Important: Note that MMapDirectory is nearly always a better choice as it uses OS caches more effectively (through memory-mapped buffers). A heap-based directory like this one can have the advantage in case of ephemeral, small, short-lived indexes when disk syncs provide an additional overhead.
type ByteBuffersDirectory struct {
}
