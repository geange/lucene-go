package util

// Bits Interface for Bitset-like structures.
type Bits interface {

	// Get Returns the value of the bit with the specified index.
	// Params: 	index – index, should be non-negative and < length(). The result of passing negative or out of
	// 			bounds values is undefined by this interface, just don't do it!
	// Returns: true if the bit is set, false otherwise.
	Get(index int) bool

	// Length Returns the number of bits in this set
	Length() int
}
