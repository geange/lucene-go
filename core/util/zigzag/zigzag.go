package zigzag

// Decode decodes a zig-zag-encoded uint64 as an int64.
//
//	Input:  {…,  5,  3,  1,  0,  2,  4,  6, …}
//	Output: {…, -3, -2, -1,  0, +1, +2, +3, …}
func Decode(x uint64) int64 {
	return int64(x>>1) ^ int64(x)<<63>>63
}

// Encode encodes an int64 as a zig-zag-encoded uint64.
//
//	Input:  {…, -3, -2, -1,  0, +1, +2, +3, …}
//	Output: {…,  5,  3,  1,  0,  2,  4,  6, …}
func Encode(x int64) uint64 {
	return uint64(x<<1) ^ uint64(x>>63)
}

// DecodeBool decodes a uint64 as a bool.
//
//	Input:  {    0,    1,    2, …}
//	Output: {false, true, true, …}
func DecodeBool(x uint64) bool {
	return x != 0
}

// EncodeBool encodes a bool as a uint64.
//
//	Input:  {false, true}
//	Output: {    0,    1}
func EncodeBool(x bool) uint64 {
	if x {
		return 1
	}
	return 0
}
