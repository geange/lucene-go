package compress

import "github.com/geange/lucene-go/core/util/array"

var _ HashTable = &HighCompressionHashTable{}

type HighCompressionHashTable struct {
	bytes      []byte
	next       int
	end        int
	hashTable  []int
	chainTable []int16
	attempts   int
}

func NewHighCompressionHashTable() *HighCompressionHashTable {
	table := &HighCompressionHashTable{
		hashTable:  make([]int, HASH_TABLE_SIZE_HC),
		chainTable: make([]int16, MAX_DISTANCE),
	}

	array.Fill(table.hashTable, -1)
	v := uint16(0xFFFF)
	array.Fill(table.chainTable, int16(v))
	return table
}

func (h *HighCompressionHashTable) Reset(bs []byte) error {

	m16 := uint16(0xFFFF)
	if h.end < len(h.chainTable) {
		// The last call to compress was done on less than 64kB, let's not reset
		// the hashTable and only reset the relevant parts of the chainTable.
		// This helps avoid slowing down calling compress() many times on short
		// inputs.
		startOffset := 0
		endOffset := 0
		if h.end != 0 {
			endOffset = (h.end-1)&MASK + 1
		}

		if startOffset < endOffset {
			array.Fill(h.chainTable[startOffset:endOffset], int16(m16))
		} else {
			array.Fill(h.chainTable[:endOffset], int16(m16))
			array.Fill(h.chainTable[startOffset:], int16(m16))
		}
	} else {
		// The last call to compress was done on a large enough amount of data
		// that it's fine to reset both tables
		array.Fill(h.hashTable, -1)
		array.Fill(h.chainTable, int16(m16))
	}

	h.bytes = bs
	h.next = 0
	h.end = len(bs)
	return nil
}

func (h *HighCompressionHashTable) InitDictionary(dictLen int) {
	for i := 0; i < dictLen; i++ {
		h.addHash(i)
	}
	h.next += dictLen
}

func (t *HighCompressionHashTable) Get(off int) (int, error) {
	for ; t.next < off; t.next++ {
		t.addHash(t.next)
	}

	v := readInt(t.bytes, off)
	h := hashHC(v)

	t.attempts = 0
	ref := t.hashTable[h]
	if ref >= off {
		// remainder from a previous call to compress()
		return -1, nil
	}

	min := off - MAX_DISTANCE + 1
	for ref >= min && t.attempts < MAX_ATTEMPTS {
		if readInt(t.bytes, ref) == v {
			return ref, nil
		}
		ref -= int(uint16(t.chainTable[ref&MASK]) & 0xFFFF)
		t.attempts++
	}
	return -1, nil
}

func (h *HighCompressionHashTable) Previous(off int) int {
	v := readInt(h.bytes, off)
	ref := off - int(uint16(h.chainTable[off&MASK])&0xFFFF)
	for ref >= 0 && h.attempts < MAX_ATTEMPTS {
		if readInt(h.bytes, ref) == v {
			return ref
		}
		ref -= int(uint16(h.chainTable[ref&MASK]) & 0xFFFF)
		h.attempts++
	}
	return -1
}

func (h *HighCompressionHashTable) addHash(off int) {
	v := readInt(h.bytes, off)
	code := hashHC(int32(v))

	delta := off - h.hashTable[code]
	if delta <= 0 || delta >= MAX_DISTANCE {
		delta = MAX_DISTANCE - 1
	}
	h.chainTable[off&MASK] = int16(delta)
	h.hashTable[code] = off
}
