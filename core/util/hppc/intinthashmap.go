package hppc

import (
	"fmt"
	"math"
)

const (
	DefaultExpectedElements = 4
	DefaultLoadFactor       = 0.75
	MinLoadFactor           = 1 / 100.0
	MaxLoadFactor           = 99 / 100.0
	MinHashArrayLength      = 4
	MaxHashArrayLength      = 1 << 30 // Max size for power of two in Go (2^30)
)

// BitMixer mimics the hashing behavior from Java's BitMixer.
type BitMixer struct{}

func (bm BitMixer) mixPhi(k int32) int32 {
	h := uint32(k) * 0x9e3779b9
	return int32(h ^ (h >> 16))
}

func (bm BitMixer) mix(n int32) int32 {
	k := uint32(n)
	k = (k ^ (k >> 16)) * 0x85ebca6b
	k = (k ^ (k >> 13)) * 0xc2b2ae35
	return int32(k ^ (k >> 16))
}

// Cursor is used to iterate over key-value pairs.
type Cursor struct {
	Index int
	Key   int32
	Value int32
}

// IntIntHashMap represents a specialized map from int to int.
type IntIntHashMap struct {
	keys        []int32
	values      []int32
	assigned    int
	mask        int
	resizeAt    int
	hasEmptyKey bool
	loadFactor  float64
	seed        int32
	bitMixer    BitMixer
}

// NewIntIntHashMap creates a new instance with default settings.
func NewIntIntHashMap() *IntIntHashMap {
	return NewIntIntHashMapWith(expected(DefaultExpectedElements), loadFactor(DefaultLoadFactor))
}

// NewIntIntHashMapWith creates a new instance with expected elements and load factor.
func NewIntIntHashMapWith(opts ...func(*IntIntHashMap)) *IntIntHashMap {
	m := &IntIntHashMap{
		loadFactor: DefaultLoadFactor,
		seed:       1, // Simple seed increment
		bitMixer:   BitMixer{},
	}
	for _, opt := range opts {
		opt(m)
	}
	m.allocateBuffers(minBufferSize(DefaultExpectedElements, m.loadFactor))
	return m
}

func expected(n int) func(*IntIntHashMap) {
	return func(m *IntIntHashMap) {
		m.assigned = 0
	}
}

func loadFactor(lf float64) func(*IntIntHashMap) {
	return func(m *IntIntHashMap) {
		if lf < MinLoadFactor || lf > MaxLoadFactor {
			panic("load factor out of range")
		}
		m.loadFactor = lf
	}
}

func (m *IntIntHashMap) allocateBuffers(size int) {
	if size < MinHashArrayLength {
		size = MinHashArrayLength
	}
	if size > MaxHashArrayLength {
		panic("maximum hash array length exceeded")
	}
	m.keys = make([]int32, size+1)
	m.values = make([]int32, size+1)
	m.mask = size - 1
	m.resizeAt = expandAtCount(size, m.loadFactor)
}

func minBufferSize(elements int, loadFactor float64) int {
	if elements < 0 {
		panic("elements must be >= 0")
	}
	length := int(math.Ceil(float64(elements) / loadFactor))
	if length == elements {
		length++
	}
	length = max(MinHashArrayLength, nextPowerOfTwo(length))
	if length > MaxHashArrayLength {
		panic("maximum array size exceeded")
	}
	return length
}

func nextPowerOfTwo(x int) int {
	if x == 0 {
		return 1
	}
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x++
	return x
}

func expandAtCount(arraySize int, loadFactor float64) int {
	capacity := int(math.Ceil(float64(arraySize) * loadFactor))
	if capacity >= arraySize {
		capacity = arraySize - 1
	}
	return capacity
}

func (m *IntIntHashMap) hashKey(key int32) int32 {
	if key == 0 {
		panic("key == 0 is handled separately")
	}
	return m.bitMixer.mixPhi(key)
}

func (m *IntIntHashMap) Put(key, value int32) int32 {
	if key == 0 {
		old := m.values[m.mask+1]
		m.values[m.mask+1] = value
		m.hasEmptyKey = true
		return old
	}

	slot := int(m.hashKey(key)) & m.mask
	for {
		existing := m.keys[slot]
		if existing == 0 {
			break
		}
		if existing == key {
			old := m.values[slot]
			m.values[slot] = value
			return old
		}
		slot = (slot + 1) & m.mask
	}

	if m.assigned >= m.resizeAt {
		m.allocateThenInsertThenRehash(slot, key, value)
	} else {
		m.keys[slot] = key
		m.values[slot] = value
	}
	m.assigned++
	return 0
}

func (m *IntIntHashMap) Get(key int32) int32 {
	if key == 0 {
		if m.hasEmptyKey {
			return m.values[m.mask+1]
		}
		return 0
	}

	slot := int(m.hashKey(key)) & m.mask
	for {
		existing := m.keys[slot]
		if existing == 0 {
			return 0
		}
		if existing == key {
			return m.values[slot]
		}
		slot = (slot + 1) & m.mask
	}
}

func (m *IntIntHashMap) GetOrDefault(key, def int32) int32 {
	if key == 0 {
		if m.hasEmptyKey {
			return m.values[m.mask+1]
		}
		return def
	}

	slot := int(m.hashKey(key)) & m.mask
	for {
		existing := m.keys[slot]
		if existing == 0 {
			return def
		}
		if existing == key {
			return m.values[slot]
		}
		slot = (slot + 1) & m.mask
	}
}

func (m *IntIntHashMap) Remove(key int32) int32 {
	if key == 0 {
		if !m.hasEmptyKey {
			return 0
		}
		val := m.values[m.mask+1]
		m.values[m.mask+1] = 0
		m.hasEmptyKey = false
		m.assigned--
		return val
	}

	slot := int(m.hashKey(key)) & m.mask
	for {
		existing := m.keys[slot]
		if existing == 0 {
			return 0
		}
		if existing == key {
			val := m.values[slot]
			m.shiftConflictingKeys(slot)
			m.assigned--
			return val
		}
		slot = (slot + 1) & m.mask
	}
}

func (m *IntIntHashMap) shiftConflictingKeys(gapSlot int) {
	distance := 0
	for {
		distance++
		slot := (gapSlot + distance) & m.mask
		existing := m.keys[slot]
		if existing == 0 {
			break
		}
		ideal := int(m.hashKey(existing)) & m.mask
		shift := (slot - ideal) & m.mask
		if shift >= distance {
			m.keys[gapSlot] = existing
			m.values[gapSlot] = m.values[slot]
			gapSlot = slot
			distance = 0
		}
	}
	m.keys[gapSlot] = 0
	m.values[gapSlot] = 0
}

func (m *IntIntHashMap) Size() int {
	if m.hasEmptyKey {
		return m.assigned + 1
	}
	return m.assigned
}

func (m *IntIntHashMap) IsEmpty() bool {
	return m.Size() == 0
}

func (m *IntIntHashMap) String() string {
	str := "["
	first := true
	for i := 0; i < len(m.keys); i++ {
		if m.keys[i] != 0 {
			if !first {
				str += ", "
			}
			str += fmt.Sprintf("%d=>%d", m.keys[i], m.values[i])
			first = false
		}
	}
	if m.hasEmptyKey {
		if !first {
			str += ", "
		}
		str += "0=>" + fmt.Sprintf("%d", m.values[m.mask+1])
	}
	str += "]"
	return str
}

func (m *IntIntHashMap) allocateThenInsertThenRehash(slot int, pendingKey, pendingValue int32) {
	newSize := nextPowerOfTwo(len(m.keys))
	if newSize > MaxHashArrayLength {
		panic("max array size exceeded")
	}
	newMap := &IntIntHashMap{}
	newMap.allocateBuffers(newSize)
	newMap.Put(pendingKey, pendingValue)

	// Rehash all entries
	for i := 0; i < len(m.keys); i++ {
		key := m.keys[i]
		if key != 0 {
			newMap.Put(key, m.values[i])
		}
	}
	*m = *newMap
}

//func max(a, b int) int {
//	if a > b {
//		return a
//	}
//	return b
//}

//func main() {
//	m := NewIntIntHashMap()
//	m.Put(1, 100)
//	m.Put(2, 200)
//	m.Put(3, 300)
//
//	fmt.Println("Size:", m.Size())
//	fmt.Println("Get(2):", m.Get(2))
//	fmt.Println("Remove(2):", m.Remove(2))
//	fmt.Println("After remove:", m)
//}
