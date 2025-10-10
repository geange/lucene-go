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
	Get(off int) (int, error)

	// Return an index that less than off and stores the same 4 bytes. Unlike get,
	// it doesn't need to be called on increasing offsets.
	// A return value of -1 indicates that no other index could be found.
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

func (f *FastCompressionHashTable) InitDictionary(dictLen int) {
	for i := 0; i < dictLen; i++ {
		v := readInt(f.bytes, i)
		h := hash(int32(v), f.hashLog)
		f.hashTable.Set(int(h), uint64(i))
	}
	f.lastOff += dictLen
}

func (f *FastCompressionHashTable) Get(off int) (int, error) {
	v := readInt(f.bytes, off)
	h := hash(int32(v), f.hashLog)

	n, err := f.hashTable.Get(int(h))
	if err != nil {
		return 0, err
	}

	ref := int(n)
	f.hashTable.Set(int(h), uint64(off))
	f.lastOff = off

	if ref < off && off-ref < MAX_DISTANCE && readInt(f.bytes, ref) == v {
		return ref, nil
	} else {
		return -1, nil
	}
}

func (f *FastCompressionHashTable) Previous(off int) int {
	return -1
}

func encodeLen(literalLen int, out store.DataOutput) error {
	for literalLen >= 0xFF {
		if err := out.WriteByte(0xFF); err != nil {
			return err
		}
		literalLen -= 0xFF
	}
	return out.WriteByte(byte(literalLen))
}

func encodeLiterals(bytes []byte, token int, out store.DataOutput) error {
	if err := out.WriteByte(byte(token)); err != nil {
		return err
	}

	literalLen := len(bytes)

	// encode literal length
	if literalLen >= 0x0F {
		encodeLen(literalLen-0x0F, out)
	}

	// encode literals
	if _, err := out.Write(bytes); err != nil {
		return err
	}
	return nil
}

func encodeLastLiterals(bytes []byte, anchor, literalLen int, out store.DataOutput) error {
	token := min(literalLen, 0x0F) << 4
	return encodeLiterals(bytes[anchor:anchor+literalLen], token, out)
}

func encodeSequence(bytes []byte, anchor, matchRef, matchOff, matchLen int, out store.DataOutput) error {
	literalLen := matchOff - anchor
	// encode token
	token := (min(literalLen, 0x0F) << 4) | min(matchLen-4, 0x0F)
	if err := encodeLiterals(bytes[anchor:anchor+literalLen], token, out); err != nil {
		return err
	}

	// encode match dec
	matchDec := matchOff - matchRef

	if err := out.WriteByte(byte(matchDec)); err != nil {
		return err
	}
	if err := out.WriteByte(byte(matchDec >> 8)); err != nil {
		return err
	}

	// encode match len
	if matchLen >= MIN_MATCH+0x0F {
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
	anchor := off

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
			matchLen := MIN_MATCH + commonBytes(bytes, ref+MIN_MATCH, off+MIN_MATCH, limit)

			// try to find a better match
			r := ht.Previous(ref)
			min := max(off-MAX_DISTANCE+1, dictOff)
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
