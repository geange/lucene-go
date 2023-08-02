package bytesref

import (
	"bytes"
	"fmt"
	"hash"
	"sort"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)

// BytesHash is a special purpose hash-map like data-structure optimized for BytesRef instances. BytesHash
// maintains mappings of byte arrays to ids (Map<BytesRef,int>) storing the hashed bytes efficiently in continuous
// storage. The mapping to the id is encapsulated inside BytesHash and is guaranteed to be increased for each
// added BytesRef.
//
// BytesHash是一种专门为[]byte实例优化的类似哈希映射的数据结构。
// BytesHash维护字节数组到id（map<[]byte, int>）的映射，有效地将散列字节存储在连续存储中。
// 到id的映射封装在BytesHash中，并保证每添加一个BytesRef都会被添加。
//
// Note: The maximum capacity BytesRef instance passed to add(BytesRef) must not be longer than
// ByteBlockPool.BYTE_BLOCK_SIZE-2. The internal storage is limited to 2GB total byte storage.
type BytesHash struct {
	sync.Mutex

	pool            *BlockPool
	bytesStart      []uint32
	hashSize        int
	hashHalfSize    int
	hashMask        uint32
	count           int
	lastCount       int
	ids             []int
	bytesStartArray StartArray
	hasher          hash.Hash32
}

const (
	DefaultCapacity = 16
)

type bytesHashOption struct {
	capacity   int
	startArray StartArray
	hasher     hash.Hash32
}

type BytesHashOption func(*bytesHashOption)

// WithCapacity capacity 需要是2的平方，如 4\16\32等
func WithCapacity(capacity int) BytesHashOption {
	return func(option *bytesHashOption) {
		option.capacity = capacity
	}
}

func WithStartArray(startArray StartArray) BytesHashOption {
	return func(option *bytesHashOption) {
		option.startArray = startArray
	}
}

func WithHash32(hasher hash.Hash32) BytesHashOption {
	return func(option *bytesHashOption) {
		option.hasher = hasher
	}
}

func NewBytesHash(pool *BlockPool, options ...BytesHashOption) (*BytesHash, error) {
	opt := &bytesHashOption{
		capacity:   DefaultCapacity,
		startArray: NewDirectStartArray(DefaultCapacity),
		hasher:     murmur3.New32WithSeed(uint32(GOOD_FAST_HASH_SEED)),
	}
	for _, option := range options {
		option(opt)
	}

	return newBytesHash(pool, opt.capacity, opt.startArray, opt.hasher), nil
}

func newBytesHash(pool *BlockPool, capacity int, startArray StartArray, hasher hash.Hash32) *BytesHash {
	bytesHash := &BytesHash{
		pool:            pool,
		bytesStart:      startArray.Init(),
		hashSize:        capacity,
		hashHalfSize:    capacity >> 1,
		hashMask:        uint32(capacity) - 1,
		count:           0,
		lastCount:       -1,
		ids:             make([]int, capacity),
		bytesStartArray: startArray,
		hasher:          hasher,
	}
	for i := range bytesHash.ids {
		bytesHash.ids[i] = -1
	}
	return bytesHash
}

// Size Returns the number of []byte/BytesRef values in this BytesHash.
// Returns: the number of BytesRef values in this BytesHash.
func (r *BytesHash) Size() int {
	return r.count
}

// Get Populates and returns a BytesRef with the bytes for the given bytesID.
// Note: the given bytesID must be a positive integer less than the current size (size())
// bytesID: the id
// ref: the BytesRef to populate
// Returns: the given BytesRef instance populated with the bytes for the given bytesID
func (r *BytesHash) Get(id int) []byte {
	return r.pool.GetBytes(r.bytesStart[id])
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

	sort.Sort(&quickSorter{
		BytesHash: r,
		ids:       compact[0:r.count],
	})
	return compact
}

type quickSorter struct {
	*BytesHash

	ids []int
}

func (r *quickSorter) Len() int {
	return len(r.ids)
}

func (r *quickSorter) Less(i, j int) bool {
	bs1 := r.pool.GetBytes(r.bytesStart[r.ids[i]])
	bs2 := r.pool.GetBytes(r.bytesStart[r.ids[j]])
	return bytes.Compare(bs1, bs2) < 0
}

func (r *quickSorter) Swap(i, j int) {
	r.ids[i], r.ids[j] = r.ids[j], r.ids[i]
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

// Add
// adds a new []byte
// bytes: the bytes to hash
// the id the given bytes are hashed if there was no mapping for the given bytes, otherwise (-(id)-1).
// This guarantees that the return value will always be >= 0 if the given bytes haven't been hashed before.
func (r *BytesHash) Add(bs []byte) (int, error) {
	length := len(bs)

	// final position
	hashPos := r.findHash(bs)
	id := r.ids[hashPos]

	if id == -1 {
		// new entry
		len2 := 2 + len(bs)
		if len2+r.pool.byteUpto > BYTE_BLOCK_SIZE {
			if len2 > BYTE_BLOCK_SIZE {
				return 0, fmt.Errorf("bytes can be at most %d in length; got %d",
					BYTE_BLOCK_SIZE-2, len(bs))
			}
			r.pool.NextBuffer()
		}
		buffer := r.pool.buffer
		byteUpto := r.pool.byteUpto
		if r.count >= len(r.bytesStart) {
			r.bytesStart = r.bytesStartArray.Grow()
		}
		id = r.count
		r.count++

		r.bytesStart[id] = uint32(byteUpto + r.pool.ByteOffset())

		// We first encode the length, followed by the
		// bytes. Len is encoded as vInt, but will consume
		// 1 or 2 bytes at most (we reject too-long terms,
		// above).

		if length < 128 {
			buffer[byteUpto] = byte(length)
			r.pool.byteUpto += length + 1
			copy(buffer[byteUpto+1:byteUpto+1+length], bs)
		} else {
			// 2 byte to store length
			buffer[byteUpto] = byte(0x80 | (length & 0x7F))
			buffer[byteUpto+1] = byte(length >> 7)
			r.pool.byteUpto += length + 2
			copy(buffer[byteUpto+2:byteUpto+2+length], bs)
		}
		r.ids[hashPos] = id

		if r.count == r.hashHalfSize {
			r.rehash(2*r.hashSize, true)
		}
		return id, nil
	}
	return -(id + 1), nil
}

func (r *BytesHash) Find(bytes []byte) int {
	return r.ids[r.findHash(bytes)]
}

func (r *BytesHash) findHash(bs []byte) int {
	code := r.doHash(bs)

	// final position
	hashPos := code & r.hashMask
	id := r.ids[hashPos]
	if id != -1 && !r.equals(id, bs) {
		// Conflict; use linear probe to find an open slot
		// (see LUCENE-5604):

		code++
		hashPos = code & r.hashMask
		id = r.ids[hashPos]

		for id != -1 && !r.equals(id, bs) {
			code++
			hashPos = code & r.hashMask
			id = r.ids[hashPos]
		}
	}
	return int(hashPos)
}

func (r *BytesHash) equals(id int, bs []byte) bool {
	textStart := r.bytesStart[id]
	blockIdx := textStart >> BYTE_BLOCK_SHIFT
	block := r.pool.buffers[blockIdx]
	pos := textStart & BYTE_BLOCK_MASK
	length, offset := uint32(0), uint32(0)
	if (block[pos] & 0x80) == 0 {
		// length is 1 byte
		length = uint32(block[pos])
		offset = pos + 1
	} else {
		// length is 2 bytes
		length = (uint32(block[pos]) & 0x7F) + ((uint32(block[pos+1]) & 0xFF) << 7)
		offset = pos + 2
	}
	return bytes.Equal(block[offset:offset+length], bs)
}

func (r *BytesHash) shrink(targetSize int) bool {
	// Cannot use ArrayUtil.shrink because we require power of 2:
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
		r.hashMask = uint32(newSize) - 1
		return true
	}

	return false
}

// AddByPoolOffset Adds a "arbitrary" int offset instead of a BytesRef term. This is used in the indexer to
// hold the hash for term vectors, because they do not redundantly store the byte[] term directly and instead
// reference the byte[] term already stored by the postings BytesHash. See add(int textStart) in
// TermsHashPerField.
func (r *BytesHash) AddByPoolOffset(offset uint32) int {
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

// ReInit
// reinitializes the BytesHash after a previous clear() call.
// If clear() has not been called previously this method has no effect.
func (r *BytesHash) ReInit() {
	if r.bytesStart == nil {
		r.bytesStart = r.bytesStartArray.Init()
	}

	if r.ids == nil {
		r.ids = make([]int, r.hashSize)
	}
}

// ByteStart
// Returns the bytesStart offset into the internally used BlockPool for the given bytesID
// Params: bytesID – the id to look up
// Returns: the bytesStart offset into the internally used BlockPool for the given id
func (r *BytesHash) ByteStart(bytesID int) uint32 {
	return r.bytesStart[bytesID]
}

// Called when hash is too small (> 50% occupied) or too large (< 20% occupied).
func (r *BytesHash) rehash(newSize int, hashOnData bool) {
	newMask := uint32(newSize - 1)
	newHash := make([]int, newSize)
	for i := 0; i < len(newHash); i++ {
		newHash[i] = -1
	}

	for i := 0; i < r.hashSize; i++ {
		id := r.ids[i]
		if id != -1 {
			code := uint32(0)
			if hashOnData {
				off := r.bytesStart[id]
				start := off & BYTE_BLOCK_MASK
				block := r.pool.buffers[off>>BYTE_BLOCK_SHIFT]
				length := uint32(0)
				pos := uint32(0)
				if block[start]&0x80 == 0 {
					// length is 1 byte
					length = uint32(block[start])
					pos = start + 1
				} else {
					length = (uint32(block[start]) & 0x7F) + (uint32(block[start+1]) << 7)
					pos = start + 2
				}
				code = r.doHash(block[pos : pos+length])
			} else {
				code = r.bytesStart[id]
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
			newHash[hashPos] = id
		}
	}

	r.hashMask = newMask
	r.ids = newHash
	r.hashSize = newSize
	r.hashHalfSize = newSize / 2
}

var (
	GOOD_FAST_HASH_SEED = time.Now().Unix()
)

func (r *BytesHash) doHash(bytes []byte) uint32 {
	r.Lock()
	defer r.Unlock()

	r.hasher.Reset()
	_, _ = r.hasher.Write(bytes)
	return r.hasher.Sum32()
}
