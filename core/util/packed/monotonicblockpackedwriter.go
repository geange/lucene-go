package packed

// MonotonicBlockPackedWriter
// A writer for large monotonically increasing sequences of positive longs.
// The sequence is divided into fixed-size blocks and for each block, values are modeled after a linear function f: x → A × x + B. The block encodes deltas from the expected values computed from this function using as few bits as possible.
// Format:
// <BLock>BlockCount
// BlockCount: ⌈ ValueCount / BlockSize ⌉
// Block: <Header, (Ints)>
// Header: <B, A, BitsPerValue>
// B: the B from f: x → A × x + B using a zig-zag encoded vLong
// A: the A from f: x → A × x + B encoded using Float.floatToIntBits(float) on 4 bytes
// BitsPerValue: a variable-length int
// Ints: if BitsPerValue is 0, then there is nothing to read and all values perfectly match the result of the function. Otherwise, these are the packed deltas from the expected value (computed from the function) using exactly BitsPerValue bits per value.
// 请参阅:
// MonotonicBlockPackedReader
// lucene.internal
type MonotonicBlockPackedWriter struct {
}
