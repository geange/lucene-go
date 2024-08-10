package index

// Bits Interface for Bitset-like structures.
type Bits interface {

	// Test
	// Returns the value of the bit with the specified index.
	// index: index, should be non-negative and < length(). The result of passing negative or out of bounds
	// values is undefined by this interface, just don't do it!
	// Returns: true if the bit is set, false otherwise.
	Test(index uint) bool

	// Len
	// Returns the number of bits in this set
	Len() uint
}
