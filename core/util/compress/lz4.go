package compress

import (
	"fmt"
	"math/bits"

	"github.com/pierrec/lz4/v4"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

const (
	MAX_ATTEMPTS = 256
)

var LZ4Compression = &LZ4{}

type LZ4 struct {
}

func (*LZ4) Compress(in []byte, out store.DataOutput) error {
	w := lz4.NewWriter(out)
	_, err := w.Write(in)
	return err
}

func (*LZ4) Decompress(in store.DataInput, out []byte) error {
	r := lz4.NewReader(in)
	_, err := r.Read(out)
	return err
}

type HashTable interface {
	// Reset this hash table in order to compress the given content.
	Reset(bs []byte) error

	// Init dictLen bytes to be used as a dictionary.
	InitDictionary(dictLen int)

	// Advance the cursor to @off and return an index that stored the same 4 bytes as b[o:o+4).
	// This may only be called on strictly increasing sequences of offsets.
	// A return value of -1 indicates that no other index could be found.
	// 查找与当前字节值相同的前一个字节的偏移量
	Get(off int) (int, error)

	// Return an index that less than off and stores the same 4 bytes. Unlike get,
	// it doesn't need to be called on increasing offsets.
	// A return value of -1 indicates that no other index could be found.
	// 查找与当前字节值相同的前一个字节的偏移量
	Previous(off int) int
}

var _ HashTable = &FastCompressionHashTable{}

type FastCompressionHashTable struct {
	bytes     []byte
	lastOff   int
	hashLog   int
	hashTable packed.Mutable
}

func NewFastCompressionHashTable() *FastCompressionHashTable {
	return &FastCompressionHashTable{}
}

func (f *FastCompressionHashTable) Reset(bs []byte) error {
	f.bytes = bs

	size := len(bs)

	bitsPerOffset, err := packed.BitsRequired(int64(size - LAST_LITERALS))
	if err != nil {
		return err
	}
	bitsPerOffsetLog := 32 - bits.LeadingZeros32(uint32(bitsPerOffset-1))
	f.hashLog = MEMORY_USAGE + 3 - bitsPerOffsetLog

	if f.hashTable == nil || f.hashTable.Size() < 1<<f.hashLog || f.hashTable.GetBitsPerValue() < bitsPerOffset {
		f.hashTable = packed.GetMutable(1<<f.hashLog, bitsPerOffset, packed.DEFAULT)
	}
	f.lastOff = -1
	return nil
}

// InitDictionary 初始化哈希表的字典部分
// 实现说明:
//
//	该函数初始化哈希表的字典部分，将字典中的每个字节添加到哈希表中。
//	字典长度为dictLen，字典起始位置为0。
func (f *FastCompressionHashTable) InitDictionary(dictLen int) {
	for i := 0; i < dictLen; i++ {
		v := readInt32(f.bytes, i)
		h := hash(v, f.hashLog)
		f.hashTable.Set(int(h), uint64(i))
	}
	f.lastOff += dictLen
}

func (f *FastCompressionHashTable) Get(off int) (int, error) {
	v := readInt32(f.bytes, off)
	h := hash(v, f.hashLog)

	n, err := f.hashTable.Get(int(h))
	if err != nil {
		return 0, err
	}

	ref := int(n)
	f.hashTable.Set(int(h), uint64(off))
	f.lastOff = off

	if ref < off && off-ref < MAX_DISTANCE && readInt32(f.bytes, ref) == v {
		return ref, nil
	} else {
		return -1, nil
	}
}

func (f *FastCompressionHashTable) Previous(off int) int {
	return -1
}

// encodeLen 将字面量序列或匹配序列的长度编码到输出流
// 实现说明:
//
//	该函数采用特殊的变长编码方式，对于大于等于0xFF的值，
//	每个字节表示0xFF的长度块，最后一个字节表示剩余的长度值。
//	这种编码方式在LZ4压缩算法中用于高效地表示可能很长的序列长度。
func encodeLen(literalLen int, out store.DataOutput) error {
	for literalLen >= 0xFF {
		if err := out.WriteByte(0xFF); err != nil {
			return err
		}
		literalLen -= 0xFF
	}
	return out.WriteByte(byte(literalLen))
}

// encodeLiterals 将字面量序列编码到输出流
// 实现说明:
//
//	该函数将字面量序列编码到输出流中。
//	当字面量长度大于等于0x0F时，需要使用变长编码，
//	否则，直接将字面量长度编码到token中。
func encodeLiterals(bytes []byte, token int, out store.DataOutput) error {
	if err := out.WriteByte(byte(token)); err != nil {
		return err
	}

	literalLen := len(bytes)

	// 当字面量长度大于等于0x0F时，需要使用变长编码
	// 否则，直接将字面量长度编码到token中
	if literalLen >= 0x0F {
		encodeLen(literalLen-0x0F, out)
	}

	// encode literals
	if _, err := out.Write(bytes); err != nil {
		return err
	}
	return nil
}

// encodeLastLiterals 将最后面的字面量序列编码到输出流
// 实现说明:
//
//	该函数将最后面的字面量序列编码到输出流中。
//	当字面量长度大于等于0x0F时，需要使用变长编码，
//	否则，直接将字面量长度编码到token中。
func encodeLastLiterals(bytes []byte, anchor, literalLen int, out store.DataOutput) error {
	token := min(literalLen, 0x0F) << 4
	return encodeLiterals(bytes[anchor:anchor+literalLen], token, out)
}

// encodeSequence 将匹配序列编码到输出流，这是LZ4压缩算法的核心部分
//
// 参数说明:
//
//	bytes: 原始输入字节数组
//	anchor: 当前序列的起始位置
//	matchRef: 匹配参考位置（之前出现过的相同序列的位置）
//	matchOff: 当前匹配序列的起始位置
//	matchLen: 匹配序列的长度
//	out: 输出流
//
// 返回值:
//
//	error: 编码过程中的错误
//
// 压缩原理:
//
//	LZ4算法基于LZ77算法变种，通过识别重复序列来实现压缩。
//	每个压缩序列由三部分组成：
//	1. Token（1字节）：高4位表示字面值长度，低4位表示匹配长度（基础值+4）
//	2. 字面值数据：如果字面值长度大于15，则需要额外的变长编码
//	3. 匹配信息：包括2字节的匹配距离和可选的匹配长度扩展
//
// 工作流程:
//  1. 计算字面值长度：当前位置到匹配起始位置的距离
//  2. 创建token：将字面值长度和匹配长度编码到一个字节中
//  3. 编码字面值部分：包括token和实际字面值数据
//  4. 计算并编码匹配距离：表示相对于当前位置，重复序列在多远的位置
//  5. 如果匹配长度超过基础值+15，则使用变长编码扩展匹配长度
func encodeSequence(bytes []byte, anchor, matchRef, matchOff, matchLen int, out store.DataOutput) error {
	// 计算字面值长度：从当前锚点到匹配开始位置的距离
	// 字面值是指无法找到匹配，需要直接存储的原始数据
	literalLen := matchOff - anchor

	// 创建token字节：
	// - 高4位存储字面值长度（如果超过15则取15，后续需要额外编码）
	// - 低4位存储匹配长度（基础值为4，所以这里是matchLen-4）
	token := (min(literalLen, 0x0F) << 4) | min(matchLen-4, 0x0F)

	// 编码字面值部分：包括token和实际的字面值数据
	// 当字面值长度大于15时，encodeLiterals函数会处理额外的长度编码
	if err := encodeLiterals(bytes[anchor:anchor+literalLen], token, out); err != nil {
		return err
	}

	// 计算匹配距离：当前匹配位置与参考位置的差值
	// 这个值表示我们需要回溯多远来找到重复的序列
	matchDec := matchOff - matchRef

	// 将匹配距离编码为2个字节（小端序）
	// 第一个字节是低8位，第二个字节是高8位
	if err := out.WriteByte(byte(matchDec)); err != nil {
		return err
	}
	if err := out.WriteByte(byte(matchDec >> 8)); err != nil {
		return err
	}

	// 处理匹配长度扩展：
	// 如果匹配长度超过MIN_MATCH+15，则需要使用变长编码存储额外长度
	// MIN_MATCH通常为4，表示最小匹配长度
	if matchLen >= MIN_MATCH+0x0F {
		// 计算额外需要编码的长度部分
		if err := encodeLen(matchLen-0x0F-MIN_MATCH, out); err != nil {
			return err
		}
	}
	return nil
}

func Compress(bytes []byte, out store.DataOutput, ht HashTable) error {
	return CompressWithDictionary(bytes, 0, 0, len(bytes), out, ht)
}

func CompressWithDictionary(bytes []byte, dictOff, dictLen, length int, out store.DataOutput, ht HashTable) error {
	if dictLen > MAX_DISTANCE {
		return fmt.Errorf("dictLen must not be greater than 64kB, but got %d", dictLen)
	}

	end := dictOff + dictLen + length

	off := dictOff + dictLen

	// anchor 记录当前匹配序列的起始位置
	anchor := off

	// 当剩余字节数大于等于LAST_LITERALS+MIN_MATCH时，
	// 可以进行匹配压缩
	if length > LAST_LITERALS+MIN_MATCH {

		limit := end - LAST_LITERALS
		matchLimit := limit - MIN_MATCH
		err := ht.Reset(bytes[dictOff : dictOff+dictLen+length])
		if err != nil {
			return err
		}
		ht.InitDictionary(dictLen)

	main:
		for off <= limit {
			// find a match
			var ref int

			for {
				if off >= matchLimit {
					break main
				}
				ref, err = ht.Get(off)
				if err != nil {
					return err
				}
				if ref != -1 {
					break
				}
				off++
			}

			// compute match length
			// 计算匹配长度
			// 注意：虽然变量名为ref，但它实际上存储的是实际的字节索引位置，而不是相对偏移量
			// 因此可以直接用作bytes数组的索引，这是LZ4算法中常见的实现方式
			// MIN_MATCH是最小匹配长度(4字节)，我们从ref+MIN_MATCH位置开始计算额外的公共字节
			// commonBytes函数会计算ref+MIN_MATCH和off+MIN_MATCH之间的公共字节数
			// 这是为了确定匹配序列的实际长度，包括MIN_MATCH个字节
			matchLen := MIN_MATCH + commonBytes(bytes, ref+MIN_MATCH, off+MIN_MATCH, limit)

			// try to find a better match
			// 尝试查找更好的匹配
			r := ht.Previous(ref)
			// 查找与当前字节值相同的前一个字节的偏移量
			// 从当前字节的偏移量开始，向回遍历哈希表，查找与当前字节值相同的前一个字节的偏移量。
			// 如果在 MAX_ATTEMPTS 次尝试后仍未找到相同字节值，则返回 -1。
			min := max(off-MAX_DISTANCE+1, dictOff) // 限制查找范围在字典长度外
			for r >= min {
				rMatchLen := MIN_MATCH + commonBytes(bytes, r+MIN_MATCH, off+MIN_MATCH, limit)
				if rMatchLen > matchLen {
					ref = r
					matchLen = rMatchLen
				}
				r = ht.Previous(r)
			}

			if err := encodeSequence(bytes, anchor, ref, off, matchLen, out); err != nil {
				return err
			}

			off += matchLen
			anchor = off
		}
	}

	// last literals
	return encodeLastLiterals(bytes, anchor, end-anchor, out)
}

// Decompress 将压缩数据解压到目标缓冲区中
// 实现说明:
//
//	该函数将压缩数据解压到目标缓冲区中。
//	解压过程中，会根据LZ4格式的token和匹配距离，
//	将压缩数据中的字面值和匹配序列解压到目标缓冲区中。
func Decompress(in store.DataInput, decompressedLen int, dest []byte) (int, error) {
	dOff := 0
	destEnd := decompressedLen

	for dOff < destEnd {
		// Read token
		token, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		tokenValue := int(token) & 0xFF

		// Process literals
		literalLen := tokenValue >> 4

		if literalLen != 0 {
			if literalLen == 0x0F {
				for {
					lenByte, err := in.ReadByte()
					if err != nil {
						return 0, err
					}
					if lenByte != 0xFF {
						literalLen += int(lenByte) & 0xFF
						break
					}
					literalLen += 0xFF
				}
			}

			n, err := in.Read(dest[dOff : dOff+literalLen])
			if err != nil {
				return 0, err
			}
			if n != literalLen {
				return 0, fmt.Errorf("short read: expected %d bytes, got %d", literalLen, n)
			}
			dOff += literalLen
		}

		if dOff >= destEnd {
			break
		}

		// Process matches
		matchDecLow, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		matchDecHigh, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		matchDec := int(matchDecLow) | (int(matchDecHigh) << 8)

		// 正确实现Lucene LZ4的匹配长度编码
		// 直接使用token的低4位作为长度增量
		matchLen := tokenValue & 0x0F

		// 处理扩展长度
		if matchLen == 0x0F {
			lenByte := byte(0)
			for {
				lenByte, err = in.ReadByte()
				if err != nil {
					return 0, err
				}
				if lenByte != 0xFF {
					break
				}
				matchLen += int(lenByte)
			}
			matchLen += int(lenByte)
		}

		// 加上MIN_MATCH得到实际匹配长度
		matchLen += MIN_MATCH

		// 确保匹配长度不超过剩余需要解压的字节数
		remain := destEnd - dOff
		if matchLen > remain {
			matchLen = remain
		}

		// 计算源引用位置
		ref := dOff - matchDec

		// 在LZ4中，matchDec 不能超过 MAX_DISTANCE
		if matchDec > MAX_DISTANCE {
			return 0, fmt.Errorf("match distance exceeds maximum allowed: %d > %d", matchDec, MAX_DISTANCE)
		}

		// 确保引用位置有效
		if ref < 0 {
			return 0, fmt.Errorf("invalid match reference: %d", ref)
		}

		// 确保目标缓冲区足够大
		end := dOff + matchLen
		if end > len(dest) {
			return 0, fmt.Errorf("destination buffer too small for matches: need %d, have %d", end, len(dest))
		}

		// 复制匹配的数据
		// 对于小长度匹配，直接逐个复制
		if matchLen < 8 || matchDec < matchLen {
			// 重叠情况或小长度 - 使用逐字节复制
			for i := 0; i < matchLen; i++ {
				dest[dOff+i] = dest[ref+i]
			}
		} else {
			// 大长度无重叠 - 使用copy函数提高性能
			copy(dest[dOff:end], dest[ref:ref+matchLen])
		}

		dOff += matchLen
	}

	return dOff, nil
}
