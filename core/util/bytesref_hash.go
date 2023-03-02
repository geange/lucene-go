package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go.uber.org/atomic"
	"sort"
	"time"

	"github.com/spaolacci/murmur3"
)

// BytesHash is a special purpose hash-map like data-structure optimized for BytesRef instances. BytesHash
// maintains mappings of byte arrays to ids (Map<BytesRef,int>) storing the hashed bytes efficiently in continuous
// storage. The mapping to the id is encapsulated inside BytesHash and is guaranteed to be increased for each
// added BytesRef.
// Note: The maximum capacity BytesRef instance passed to add(BytesRef) must not be longer than
// ByteBlockPool.BYTE_BLOCK_SIZE-2. The internal storage is limited to 2GB total byte storage.
type BytesHash struct {
	//BASE_RAM_BYTES   int
	//DEFAULT_CAPACITY int

	pool            *ByteBlockPool
	bytesStart      []int
	hashSize        int
	hashHalfSize    int
	hashMask        int
	count           int
	lastCount       int
	ids             []int
	bytesStartArray BytesStartArray
}

const (
	DEFAULT_CAPACITY = 16
)

func NewBytesRefHash(pool *ByteBlockPool) *BytesHash {
	return NewBytesRefHashV1(pool, 16, NewDirectBytesStartArray(16))
}

func NewBytesRefHashV1(pool *ByteBlockPool, capacity int, bytesStartArray BytesStartArray) *BytesHash {
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
	return hash
}

func newBytesRefHash() *BytesHash {
	return &BytesHash{
		//BASE_RAM_BYTES:   binary.Size(BytesHash{}) + 8,
		//DEFAULT_CAPACITY: 16,
		pool:            nil,
		bytesStart:      make([]int, 0),
		hashSize:        0,
		hashHalfSize:    0,
		hashMask:        0,
		count:           0,
		lastCount:       -1,
		ids:             make([]int, 0),
		bytesStartArray: nil,
	}
}

// Size Returns the number of BytesRef values in this BytesHash.
// Returns: the number of BytesRef values in this BytesHash.
func (r *BytesHash) Size() int {
	return r.count
}

// Get Populates and returns a BytesRef with the bytes for the given bytesID.
// Note: the given bytesID must be a positive integer less than the current size (size())
// Params: 	bytesID – the id
//
//	ref – the BytesRef to populate
//
// Returns: the given BytesRef instance populated with the bytes for the given bytesID
func (r *BytesHash) Get(bytesID int) []byte {
	return r.pool.GetBytes(r.bytesStart[bytesID])
}

// Compact Returns the ids array in arbitrary order. Valid ids start at offset of 0 and end at a limit of size() - 1
// Note: This is a destructive operation. clear() must be called in order to reuse this BytesHash instance.
func (r *BytesHash) Compact() []int {
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
// Note: This is a destructive operation. clear() must be called in order to reuse this BytesHash instance.
func (r *BytesHash) Sort() []int {
	compact := r.Compact()

	sorter := &RadixSorter{
		Ids:       compact[0:r.count],
		BytesHash: r,
	}
	sort.Sort(sorter)

	return compact
}

type RadixSorter struct {
	Ids []int
	*BytesHash
}

func (r *RadixSorter) Len() int {
	return len(r.Ids)
}

func (r *RadixSorter) Less(i, j int) bool {
	bytes1 := r.pool.get(r.bytesStart[r.Ids[i]])
	bytes2 := r.pool.get(r.bytesStart[r.Ids[j]])
	return bytes.Compare(bytes1, bytes2) < 0
}

func (r *RadixSorter) Swap(i, j int) {
	r.Ids[i], r.Ids[j] = r.Ids[j], r.Ids[i]
}

func (r *BytesHash) Clear(resetPool bool) {
	r.lastCount = r.count
	r.count = 0
	if resetPool {
		r.pool.Reset(false, false)
	}
	r.bytesStart = r.bytesStartArray.Clear()
	if r.lastCount != -1 && r.shrink(r.lastCount) {
		// shrink clears the hash entries
		return
	}

	for i := 0; i < len(r.ids); i++ {
		r.ids[i] = -1
	}
}

func (r *BytesHash) Close() {
	r.Clear(true)
	r.ids = nil
}

// Add Adds a new BytesRef
// Params: bytes – the bytes to hash
// Returns: the id the given bytes are hashed if there was no mapping for the given bytes, otherwise (-(id)-1).
// This guarantees that the return value will always be >= 0 if the given bytes haven't been hashed before.
// Throws: BytesHash.MaxBytesLengthExceededException – if the given bytes are > 2 + ByteBlockPool.BYTE_BLOCK_SIZE
func (r *BytesHash) Add(bytes []byte) (int, error) {
	length := len(bytes)

	// final position
	hashPos := r.findHash(bytes)
	e := r.ids[hashPos]

	if e == -1 {
		// new entry
		len2 := 2 + len(bytes)
		if len2+r.pool.byteUpto > BYTE_BLOCK_SIZE {
			if len2 > BYTE_BLOCK_SIZE {
				return 0, fmt.Errorf("bytes can be at most %d in length; got %d",
					BYTE_BLOCK_SIZE-2, len(bytes))
			}
			r.pool.NextBuffer()
		}
		buffer := r.pool.buffer
		bufferUpto := r.pool.bufferUpto
		if r.count >= len(r.bytesStart) {
			r.bytesStart = r.bytesStartArray.Grow()
		}
		e = r.count
		r.count++

		r.bytesStart[e] = bufferUpto + r.pool.ByteOffset

		// We first encode the length, followed by the
		// bytes. Length is encoded as vInt, but will consume
		// 1 or 2 bytes at most (we reject too-long terms,
		// above).

		if length < 128 {
			buffer[bufferUpto] = byte(length)
			r.pool.bufferUpto += length + 1
			copy(buffer[bufferUpto+1:bufferUpto+1+length], bytes)
		} else {
			// 2 byte to store length
			buffer[bufferUpto] = byte(0x80 | (length & 0x7f))
			buffer[bufferUpto+1] = byte((length >> 7) & 0xff)
			r.pool.bufferUpto += length + 2
			copy(buffer[bufferUpto+2:bufferUpto+2+length], bytes)
		}
		r.ids[hashPos] = e

		if r.count == r.hashHalfSize {
			r.rehash(2*r.hashSize, true)
		}
		return e, nil
	}
	return -(e + 1), nil
}

func (r *BytesHash) Find(bytes []byte) int {
	return r.ids[r.findHash(bytes)]
}

func (r *BytesHash) findHash(bytes []byte) int {
	code := r.doHash(bytes)

	// final position
	hashPos := code & r.hashMask
	e := r.ids[hashPos]
	if e != -1 && !r.equals(e, bytes) {
		// Conflict; use linear probe to find an open slot
		// (see LUCENE-5604):

		code++
		hashPos = code & r.hashMask
		e = r.ids[hashPos]

		for e != -1 && !r.equals(e, bytes) {
			code++
			hashPos = code & r.hashMask
			e = r.ids[hashPos]
		}
	}
	return hashPos
}

func (r *BytesHash) equals(id int, b []byte) bool {
	textStart := r.bytesStart[id]
	array := r.pool.buffers[textStart>>BYTE_BLOCK_SHIFT]
	pos := textStart & BYTE_BLOCK_MASK
	length, offset := 0, 0
	if (array[pos] & 0x80) == 0 {
		// length is 1 byte
		length = int(array[pos])
		offset = pos + 1
	} else {
		// length is 2 bytes
		value, n := binary.Uvarint(array[pos:])
		length = int(value)
		offset += n
	}
	return bytes.Equal(array[offset:offset+length], b)
}

func (r *BytesHash) shrink(targetSize int) bool {
	// Cannot use ArrayUtil.shrink because we require power
	// of 2:
	newSize := r.hashSize
	for newSize >= 8 && newSize/4 > targetSize {
		newSize /= 2
	}

	if newSize != r.hashSize {
		r.hashSize = newSize
		r.ids = make([]int, newSize)
		for i := 0; i < len(r.ids); i++ {
			r.ids[i] = -1
		}
		r.hashHalfSize = newSize / 2
		r.hashMask = newSize - 1
		return true
	}

	return false
}

// AddByPoolOffset Adds a "arbitrary" int offset instead of a BytesRef term. This is used in the indexer to
// hold the hash for term vectors, because they do not redundantly store the byte[] term directly and instead
// reference the byte[] term already stored by the postings BytesHash. See add(int textStart) in
// TermsHashPerField.
func (r *BytesHash) AddByPoolOffset(offset int) int {
	// final position
	code := offset
	hashPos := offset & r.hashMask
	e := r.ids[hashPos]
	if e != -1 && r.bytesStart[e] != offset {
		// Conflict; use linear probe to find an open slot
		// (see LUCENE-5604):
		code++
		hashPos = code & r.hashMask
		e = r.ids[hashPos]

		for e != -1 && r.bytesStart[e] != offset {
			code++
			hashPos = code & r.hashMask
			e = r.ids[hashPos]
		}
	}

	if e == -1 {
		// new entry
		if r.count >= len(r.bytesStart) {
			r.bytesStart = r.bytesStartArray.Grow()
		}
		e = r.count
		r.count++
		r.bytesStart[e] = offset
		r.ids[hashPos] = e

		if r.count == r.hashHalfSize {
			r.rehash(2*r.hashSize, false)
		}
		return e
	}
	return -(e + 1)
}

// Reinit reinitializes the BytesHash after a previous clear() call. If clear() has not been called previously this method has no effect.
func (r *BytesHash) Reinit() {
	if r.bytesStart == nil {
		r.bytesStart = r.bytesStartArray.Init()
	}

	if r.ids == nil {
		r.ids = make([]int, r.hashSize)
	}
}

// ByteStart Returns the bytesStart offset into the internally used ByteBlockPool for the given bytesID
// Params: bytesID – the id to look up
// Returns: the bytesStart offset into the internally used ByteBlockPool for the given id
func (r *BytesHash) ByteStart(bytesID int) int {
	return r.bytesStart[bytesID]
}

// Called when hash is too small (> 50% occupied) or too large (< 20% occupied).
func (r *BytesHash) rehash(newSize int, hashOnData bool) {
	newMask := newSize - 1
	newHash := make([]int, newSize)
	for i := 0; i < len(newHash); i++ {
		newHash[i] = -1
	}

	for i := 0; i < r.hashSize; i++ {
		e0 := r.ids[i]
		if e0 != -1 {
			code := 0
			if hashOnData {
				off := r.bytesStart[e0]
				start := off & BYTE_BLOCK_MASK
				bytes := r.pool.buffers[off>>BYTE_BLOCK_SHIFT]
				length := 0
				pos := 0
				if bytes[start]&0x80 == 0 {
					// length is 1 byte
					length = int(bytes[start])
					pos = start + 1
				} else {
					length = int((bytes[start] & 0x7f) + ((bytes[start+1] & 0xff) << 7))
					pos = start + 2
				}
				code = r.doHash(bytes[pos : pos+length])
			} else {
				code = r.bytesStart[e0]
			}

			hashPos := code & newMask
			if newHash[hashPos] != -1 {
				code++
				hashPos = code & newMask
				for newHash[hashPos] != -1 {
					code++
					hashPos = code & newMask
				}
			}
			newHash[hashPos] = e0
		}
	}
}

var (
	GOOD_FAST_HASH_SEED = time.Now().Unix()
)

func (r *BytesHash) doHash(bytes []byte) int {
	hasher := murmur3.New32WithSeed(uint32(GOOD_FAST_HASH_SEED))
	_, _ = hasher.Write(bytes)
	return int(hasher.Sum32())
}

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
