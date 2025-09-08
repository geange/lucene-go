package compress

import (
	"encoding/binary"
	"math/bits"

	"github.com/pierrec/lz4/v4"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
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

func (f *FastCompressionHashTable) Reset(bs []byte) error {
	f.bytes = bs

	size := len(bs)

	bitsPerOffset, err := packed.BitsRequired(int64(size))
	if err != nil {
		return err
	}
	bitsPerOffsetLog := bits.LeadingZeros32(uint32(bitsPerOffset - 1))
	f.hashLog = MEMORY_USAGE + 3 - bitsPerOffsetLog

	if f.hashTable == nil || f.hashTable.Size() < 1<<f.hashLog || f.hashTable.GetBitsPerValue() < bitsPerOffset {
		f.hashTable = packed.GetMutable(1<<f.hashLog, bitsPerOffset, packed.DEFAULT)
	}
	f.lastOff = -1
	return nil
}

func (f *FastCompressionHashTable) InitDictionary(dictLen int) {
	for i := 0; i < dictLen; i++ {
		v := binary.BigEndian.Uint32(f.bytes[i:])
		h := hashInt32(int32(v), f.hashLog)
		f.hashTable.Set(int(h), uint64(i))
	}
	f.lastOff += dictLen
}

func hashInt32(i int32, hashBits int) uint32 {
	return (uint32(i*-1640531535) >> (32 - hashBits))
}

func (f *FastCompressionHashTable) Get(off int) (int, error) {
	v := binary.BigEndian.Uint32(f.bytes[off:])
	h := hashInt32(int32(v), f.hashLog)

	n, err := f.hashTable.Get(int(h))
	if err != nil {
		return 0, err
	}

	ref := int(n)
	f.hashTable.Set(int(h), uint64(off))
	f.lastOff = off

	if ref < off && off-ref < MAX_DISTANCE && binary.BigEndian.Uint32(f.bytes[ref:]) == v {
		return ref, nil
	} else {
		return -1, nil
	}
}

func (f *FastCompressionHashTable) Previous(off int) int {
	return -1
}

var _ HashTable = &HighCompressionHashTable{}

type HighCompressionHashTable struct {
}

func (f *HighCompressionHashTable) Reset(bs []byte) error {
	panic("")
}

func (f *HighCompressionHashTable) InitDictionary(dictLen int) {

}

func (f *HighCompressionHashTable) Get(off int) (int, error) {
	panic("")
}

func (f *HighCompressionHashTable) Previous(off int) int {
	panic("")
}
