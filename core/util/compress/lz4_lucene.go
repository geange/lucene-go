package compress

import "github.com/geange/lucene-go/core/util/array"

const (
	MEMORY_USAGE       = 14
	MIN_MATCH          = 4       // minimum length of a match
	MAX_DISTANCE       = 1 << 16 // maximum distance of a reference
	LAST_LITERALS      = 5       // the last 5 bytes must be encoded as literals
	HASH_LOG_HC        = 15      // log size of the dictionary for compressHC
	HASH_TABLE_SIZE_HC = 1 << HASH_LOG_HC
)

func (*LZ4) hash(i int32, hashBits int) int32 {
	return int32(uint32(i*-1640531535) >> (32 - hashBits))
}

func (l *LZ4) hashHC(i int32) int32 {
	return l.hash(i, HASH_LOG_HC)
}

func (l *LZ4) readInt(buf []byte, i int) int32 {
	n := (uint32(buf[i]) << 24) | (uint32(buf[i+1]) << 16) | (uint32(buf[i+2]) << 8) | (uint32(buf[i+3]) & 0xFF)
	return int32(n)
}

func (l *LZ4) commonBytes(b []byte, o1, o2, limit int) int {
	return array.Mismatch(b[o1:o1+limit], b[o2:o2+limit])
}
