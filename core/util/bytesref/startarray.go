package bytesref

// StartArray
// Manages allocation of the per-term addresses.
type StartArray interface {
	// Init Initializes the StartArray. This call will allocate memory
	// Returns: the initialized bytes start array
	Init() []uint32

	// Grow Grows the BytesHash.BytesStartArray
	// Returns: the grown array
	Grow() []uint32

	// Clear clears the BytesHash.BytesStartArray and returns the cleared instance.
	// Returns: the cleared instance, this might be null
	Clear() []uint32

	// BytesUsed A Counter reference holding the number of bytes used by this BytesHash.BytesStartArray.
	// The BytesHash uses this reference to track it memory usage
	// Returns: a AtomicLong reference holding the number of bytes used by this BytesHash.BytesStartArray.
	// BytesUsed() *atomic.Int64
}

// DirectStartArray
// A simple BytesHash.BytesStartArray that tracks memory allocation using a
// private Counter instance.
type DirectStartArray struct {
	// TODO: can't we just merge this
	// TrackingDirectBytesStartArray...?  Just add a ctor
	// that makes a private bytesUsed?
	initSize   int
	bytesStart []uint32
}

func NewDirectStartArray(initSize int) *DirectStartArray {
	return &DirectStartArray{
		initSize:   initSize,
		bytesStart: nil,
	}
}

func (d *DirectStartArray) Init() []uint32 {
	d.bytesStart = make([]uint32, d.initSize)
	return d.bytesStart
}

func (d *DirectStartArray) Grow() []uint32 {
	d.bytesStart = append(d.bytesStart, 0)
	return d.bytesStart
}

func (d *DirectStartArray) Clear() []uint32 {
	d.bytesStart = d.bytesStart[:0]
	return d.bytesStart
}
