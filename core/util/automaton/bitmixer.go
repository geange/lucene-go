package automaton

const (
	// Golden ratio bit mixers.
	PHI_C32 = uint32(0x9e3779b9)
	PHI_C64 = uint64(0x9e3779b97f4a7c15)
)

func mix(key int) int {
	return mix32(key)
}

// MurmurHash3算法中的32位最终混合步骤
func mix32(v int) int {
	k := uint32(v)
	k = (k ^ (k >> 16)) * 0x85ebca6b
	k = (k ^ (k >> 13)) * 0xc2b2ae35
	return int(k ^ (k >> 16))
}

//func mixPhi(k int32) int32 {
//	h := k * int32(PHI_C32)
//	return (h) ^ int32(uint32(h)>>16)
//}
