package util

import (
	"encoding/binary"
	"go.uber.org/atomic"
)

// BytesRefHash is a special purpose hash-map like data-structure optimized for BytesRef instances. BytesRefHash
// maintains mappings of byte arrays to ids (Map<BytesRef,int>) storing the hashed bytes efficiently in continuous
// storage. The mapping to the id is encapsulated inside BytesRefHash and is guaranteed to be increased for each
// added BytesRef.
// Note: The maximum capacity BytesRef instance passed to add(BytesRef) must not be longer than
// ByteBlockPool.BYTE_BLOCK_SIZE-2. The internal storage is limited to 2GB total byte storage.
type BytesRefHash struct {
	BASE_RAM_BYTES   int
	DEFAULT_CAPACITY int

	pool            *ByteBlockPool
	bytesStart      []int
	hashSize        int
	hashHalfSize    int
	hashMask        int
	count           int
	lastCount       int
	ids             []int
	bytesStartArray BytesStartArray
	bytesUsed       *atomic.Int64
}

func NewBytesRefHash(pool *ByteBlockPool) *BytesRefHash {
	return NewBytesRefHashV1(pool, 16, NewDirectBytesStartArray(16))
}

func NewBytesRefHashV1(pool *ByteBlockPool, capacity int, bytesStartArray BytesStartArray) *BytesRefHash {
	hash := newBytesRefHash()
	hash.hashSize = capacity
	hash.hashHalfSize = hash.hashSize >> 1
	hash.hashMask = hash.hashSize - 1
	hash.pool = pool
	hash.ids = make([]int, hash.hashSize)
	for i := range hash.ids {
		hash.ids[i] = -1
	}
	hash.bytesStartArray = bytesStartArray
	hash.bytesStart = bytesStartArray.Init()
	hash.bytesUsed = bytesStartArray.BytesUsed()
	hash.bytesUsed.Add(int64(hash.hashSize * 8))
	return hash
}

func newBytesRefHash() *BytesRefHash {
	return &BytesRefHash{
		BASE_RAM_BYTES:   binary.Size(BytesRefHash{}) + 8,
		DEFAULT_CAPACITY: 16,
		pool:             nil,
		bytesStart:       make([]int, 0),
		hashSize:         0,
		hashHalfSize:     0,
		hashMask:         0,
		count:            0,
		lastCount:        -1,
		ids:              make([]int, 0),
		bytesStartArray:  nil,
		bytesUsed:        atomic.NewInt64(0),
	}
}

// Size Returns the number of BytesRef values in this BytesRefHash.
// Returns: the number of BytesRef values in this BytesRefHash.
func (r *BytesRefHash) Size() int {
	return r.count
}

// Get Populates and returns a BytesRef with the bytes for the given bytesID.
// Note: the given bytesID must be a positive integer less than the current size (size())
// Params: 	bytesID – the id
//			ref – the BytesRef to populate
// Returns: the given BytesRef instance populated with the bytes for the given bytesID
func (r *BytesRefHash) Get(bytesID int, ref *BytesRef) *BytesRef {
	r.pool.SetBytesRefV2(ref, r.bytesStart[bytesID])
	return ref
}

// Compact Returns the ids array in arbitrary order. Valid ids start at offset of 0 and end at a limit of size() - 1
// Note: This is a destructive operation. clear() must be called in order to reuse this BytesRefHash instance.
func (r *BytesRefHash) Compact() []int {
	upto := 0
	for i := 0; i < r.hashSize; i++ {
		if r.ids[i] != -1 {
			if upto < i {
				r.ids[upto] = r.ids[i]
				r.ids[i] = -1
			}
			upto++
		}
	}

	r.lastCount = r.count
	return r.ids
}

// Sort Returns the values array sorted by the referenced byte values.
// Note: This is a destructive operation. clear() must be called in order to reuse this BytesRefHash instance.
func (r *BytesRefHash) Sort() []int {
	compact := r.Compact()
	panic("")
	// TODO
	return compact
}

func (r *BytesRefHash) Add(bytes []byte) int {
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
