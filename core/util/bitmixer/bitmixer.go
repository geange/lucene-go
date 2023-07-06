package bitmixer

func Mix32(v int) int {
	k := uint32(v)
	k = (k ^ (k >> 16)) * 0x85ebca6b
	k = (k ^ (k >> 13)) * 0xc2b2ae35
	return int(k ^ (k >> 16))
}
