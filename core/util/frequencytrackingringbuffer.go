package util

// FrequencyTrackingRingBuffer
// A ring buffer that tracks the frequency of the integers that it contains. This is typically useful to track the hash codes of popular recently-used items. This data-structure requires 22 bytes per entry on average (between 16 and 28).
// lucene. internal
type FrequencyTrackingRingBuffer struct {
	maxSize     int
	buffer      []int32
	position    int
	frequencies *intBag
}

func NewFrequencyTrackingRingBuffer(maxSize int, sentinel int32) *FrequencyTrackingRingBuffer {
	fb := &FrequencyTrackingRingBuffer{
		maxSize:     maxSize,
		buffer:      make([]int32, maxSize),
		position:    0,
		frequencies: newIntBag(maxSize),
	}

	for i := range fb.buffer {
		fb.buffer[i] = sentinel
	}

	for i := 0; i < maxSize; i++ {
		fb.frequencies.add(sentinel)
	}

	return fb
}

// Add a new item to this ring buffer, potentially removing the oldest entry from this buffer if it is already full.
func (f *FrequencyTrackingRingBuffer) Add(i int32) {
	// remove the previous value
	//removed := f.buffer[f.position]
	//removedFromBag := f.frequencies.remove(removed)
	// add the new value
	f.buffer[f.position] = i
	f.frequencies.add(i)
	// increment the position
	f.position++
	if f.position == f.maxSize {
		f.position = 0
	}
}

// Frequency
// Returns the frequency of the provided key in the ring buffer.
func (f *FrequencyTrackingRingBuffer) Frequency(key int32) int32 {
	return f.frequencies.frequency(key)
}

type intBag struct {
	keys  []int32
	freqs []int32
	mask  int32
}

func newIntBag(maxSize int) *intBag {
	// load factor of 2/3
	capacity := int32(max(2, maxSize*3/2))
	// round up to the next power of two
	capacity = HighestOneBit(capacity-1) << 1

	return &intBag{
		keys:  make([]int32, capacity),
		freqs: make([]int32, capacity),
		mask:  capacity - 1,
	}
}

// Return the frequency of the give key in the bag.
func (i *intBag) frequency(key int32) int32 {
	for slot := key & i.mask; ; slot++ {
		if i.keys[slot] == key {
			return i.freqs[slot]
		} else if i.freqs[slot] == 0 {
			return 0
		}
	}
}

// Increment the frequency of the given key by 1 and return its new frequency.
func (i *intBag) add(key int32) int32 {
	for slot := key & i.mask; ; slot = (slot + 1) & i.mask {
		if i.freqs[slot] == 0 {
			i.keys[slot] = key
			i.freqs[slot] = 1
			return i.freqs[slot]
		} else if i.keys[slot] == key {
			i.freqs[slot]++
			return i.freqs[slot]
		}
	}
}

// Decrement the frequency of the given key by one, or do nothing if the key is not present in the bag. Returns true iff the key was contained in the bag.
func (i *intBag) remove(key int32) bool {
	for slot := key & i.mask; ; slot = (slot + 1) & i.mask {
		if i.freqs[slot] == 0 {
			// no such key in the bag
			return false
		} else if i.keys[slot] == key {
			i.freqs[slot]--
			newFreq := i.freqs[slot]
			if newFreq == 0 { // removed
				i.relocateAdjacentKeys(slot)
			}
			return true
		}
	}
}

func (i *intBag) relocateAdjacentKeys(freeSlot int32) {
	for slot := (freeSlot + 1) & i.mask; ; slot = (slot + 1) & i.mask {
		freq := i.freqs[slot]
		if freq == 0 {
			// end of the collision chain, we're done
			break
		}
		key := i.keys[slot]
		// the slot where <code>key</code> should be if there were no collisions
		expectedSlot := key & i.mask
		// if the free slot is between the expected slot and the slot where the
		// key is, then we can relocate there
		if between(expectedSlot, slot, freeSlot) {
			i.keys[freeSlot] = key
			i.freqs[freeSlot] = freq
			// slot is the new free slot
			i.freqs[slot] = 0
			freeSlot = slot
		}
	}
}

// Given a chain of occupied slots between chainStart and chainEnd, return whether slot is between the start and end of the chain.
func between(chainStart, chainEnd, slot int32) bool {
	if chainStart <= chainEnd {
		return chainStart <= slot && slot <= chainEnd
	} else {
		// the chain is across the end of the array
		return slot >= chainStart || slot <= chainEnd
	}
}
