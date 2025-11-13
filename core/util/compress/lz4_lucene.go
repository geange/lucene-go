package compress

import (
	"encoding/binary"

	"github.com/geange/lucene-go/core/util/array"
)

const (
	MEMORY_USAGE       = 14
	MIN_MATCH          = 4       // minimum length of a match
	MAX_DISTANCE       = 1 << 16 // maximum distance of a reference
	MASK               = MAX_DISTANCE - 1
	LAST_LITERALS      = 5  // the last 5 bytes must be encoded as literals
	HASH_LOG_HC        = 15 // log size of the dictionary for compressHC
	HASH_TABLE_SIZE_HC = 1 << HASH_LOG_HC
)

func hash(i int32, hashBits int) int32 {
	return int32(uint32(i*-1640531535) >> (32 - hashBits))
}

func hashHC(i int32) int32 {
	return hash(i, HASH_LOG_HC)
}

func readInt32(buf []byte, i int) int32 {
	n := binary.BigEndian.Uint32(buf[i:])
	return int32(n)
}

// commonBytes 计算两个字节数组从指定偏移量开始的公共字节数
// 实现说明:
//
//	该函数计算两个字节数组从指定偏移量开始的公共字节数。
//	它从 o1 开始，比较 b[o1:limit] 和 b[o2:limit] 中的字节，
//	返回第一个不匹配的字节的索引减去 o1。
//	如果两个数组在 limit 之前完全匹配，则返回 limit-o1。
func commonBytes(b []byte, o1, o2, limit int) int {
	return array.Mismatch(b[o1:limit], b[o2:limit])
}
