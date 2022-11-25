package fst

var BitTable = &bitTable{}

// bitTable Helper methods to read the bit-table of a direct addressing node. Only valid for Arc
// with Arc.nodeFlags() == ARCS_FOR_DIRECT_ADDRESSING.
type bitTable struct {
}

// See BitTableUtil.isBitSet(int, FST.BytesReader).
func (b *bitTable) isBitSet(bitIndex int, arc *Arc, in BytesReader) (bool, error) {
	err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return false, err
	}
	in.SetPosition(arc.bitTableStart)
	return isBitSet(bitIndex, in)
}

// See BitTableUtil.countBits(int, FST.BytesReader).
// The count of bit set is the number of arcs of a direct addressing node.
func (b *bitTable) countBits(arc *Arc, in BytesReader) (int64, error) {
	err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return 0, err
	}
	in.SetPosition(arc.bitTableStart)

	numPresenceBytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return 0, err
	}
	return countBits(numPresenceBytes, in)
}

// See BitTableUtil.countBitsUpTo(int, FST.BytesReader).
func (b *bitTable) countBitsUpTo(bitIndex int, arc *Arc, in BytesReader) (int, error) {
	err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return 0, err
	}
	in.SetPosition(arc.bitTableStart)
	return countBitsUpTo(bitIndex, in)
}

// See BitTableUtil.nextBitSet(int, int, FST.BytesReader).
func (b *bitTable) nextBitSet(bitIndex int, arc *Arc, in BytesReader) (int, error) {
	err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return 0, err
	}
	in.SetPosition(arc.bitTableStart)

	bytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return 0, err
	}
	return nextBitSet(bitIndex, int(bytes), in)
}

// See BitTableUtil.previousBitSet(int, FST.BytesReader).
func (b *bitTable) previousBitSet(bitIndex int, arc *Arc, in BytesReader) (int, error) {
	err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return 0, err
	}
	in.SetPosition(arc.bitTableStart)
	return previousBitSet(bitIndex, in)
}

// Asserts the bit-table of the provided FST.Arc is valid.
func (b *bitTable) assertIsValid(arc *Arc, in BytesReader) (bool, error) {
	err := assert(arc.BytesPerArc() > 0)
	if err != nil {
		return false, err
	}
	err = assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return false, err
	}
	// First bit must be set.
	ok, err := b.isBitSet(0, arc, in)
	if err != nil {
		return false, err
	}
	err = assert(ok)
	if err != nil {
		return false, err
	}
	// Last bit must be set.
	ok, err = b.isBitSet(int(arc.NumArcs()-1), arc, in)
	if err != nil {
		return false, err
	}
	err = assert(ok)
	if err != nil {
		return false, err
	}
	// No bit set after the last arc.
	bitSet, err := b.nextBitSet(int(arc.NumArcs()-1), arc, in)
	if err != nil {
		return false, err
	}
	err = assert(bitSet == -1)
	if err != nil {
		return false, err
	}
	return true, nil
}
