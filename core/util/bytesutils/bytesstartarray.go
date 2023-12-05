package bytesutils

// BytesStartArray Manages allocation of the per-term addresses.
type BytesStartArray interface {
	// Init Initializes the BytesStartArray. This call will allocate memory
	// Returns: the initialized bytes start array
	Init() []int

	// Grow Grows the BytesHash.BytesStartArray
	// Returns: the grown array
	Grow() []int

	// Clear clears the BytesHash.BytesStartArray and returns the cleared instance.
	// Returns: the cleared instance, this might be null
	Clear() []int

	// BytesUsed A Counter reference holding the number of bytes used by this BytesHash.BytesStartArray.
	// The BytesHash uses this reference to track it memory usage
	// Returns: a AtomicLong reference holding the number of bytes used by this BytesHash.BytesStartArray.
	// BytesUsed() *atomic.Int64
}

// DirectBytesStartArray A simple BytesHash.BytesStartArray that tracks memory allocation using a
// private Counter instance.
type DirectBytesStartArray struct {
	// TODO: can't we just merge this
	// TrackingDirectBytesStartArray...?  Just add a ctor
	// that makes a private bytesUsed?
	initSize   int
	bytesStart []int
}

func NewDirectBytesStartArray(initSize int) *DirectBytesStartArray {
	return &DirectBytesStartArray{
		initSize:   initSize,
		bytesStart: nil,
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
	d.bytesStart = d.bytesStart[:0]
	return d.bytesStart
}
