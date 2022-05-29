package core

import "go.uber.org/atomic"

// BytesRefHash is a special purpose hash-map like data-structure optimized for BytesRef instances.
// BytesRefHash maintains mappings of byte arrays to ids (Map<BytesRef,int>) storing the hashed bytes
// efficiently in continuous storage. The mapping to the id is encapsulated inside BytesRefHash and
// is guaranteed to be increased for each added BytesRef.
// Note: The maximum capacity BytesRef instance passed to add(BytesRef) must not be longer than
// ByteBlockPool.BYTE_BLOCK_SIZE-2. The internal storage is limited to 2GB total byte storage.
type BytesRefHash struct {
}

func (b *BytesRefHash) Add(bytes []byte) int {
	panic("")
}

type BytesStartArray interface {
	// Init Initializes the BytesStartArray. This call will allocate memory
	// Returns: the initialized bytes start array
	Init() []int

	// Grow Grows the BytesRefHash.BytesStartArray
	// Returns: the grown array
	Grow() []int

	// Clear clears the BytesRefHash.BytesStartArray and returns the cleared instance.
	// Returns: the cleared instance, this might be null
	Clear() []int

	// BytesUsed A Counter reference holding the number of bytes used by this BytesRefHash.BytesStartArray.
	// The BytesRefHash uses this reference to track it memory usage
	// Returns: a AtomicLong reference holding the number of bytes used by this BytesRefHash.BytesStartArray.
	BytesUsed() *atomic.Int64
}

// DirectBytesStartArray A simple BytesRefHash.BytesStartArray that tracks memory allocation using a
// private Counter instance.
type DirectBytesStartArray struct {
	// TODO: can't we just merge this w/
	// TrackingDirectBytesStartArray...?  Just add a ctor
	// that makes a private bytesUsed?
	initSize   int
	bytesStart []int
	bytesUsed  *atomic.Int64
}

func NewDirectBytesStartArray(initSize int) *DirectBytesStartArray {
	return &DirectBytesStartArray{
		initSize:   initSize,
		bytesStart: nil,
		bytesUsed:  atomic.NewInt64(0),
	}
}

func (d *DirectBytesStartArray) Init() []int {
	d.bytesStart = make([]int, d.initSize)
	return d.bytesStart
}

func (d *DirectBytesStartArray) Grow() []int {
	d.bytesStart = append(d.bytesStart, 0)
	return d.bytesStart
}

func (d *DirectBytesStartArray) Clear() []int {
	d.bytesStart = nil
	return nil
}

func (d *DirectBytesStartArray) BytesUsed() *atomic.Int64 {
	return d.bytesUsed
}
