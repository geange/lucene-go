package fst

// bitTable Helper methods to read the bit-table of a direct addressing node. Only valid for Arc
// with Arc.nodeFlags() == ARCS_FOR_DIRECT_ADDRESSING.
type bitTable struct {
}

// IsBitSet See BitTableUtil.IsBitSet(int, Fst.BytesReader).
func IsBitSet[T any](bitIndex int, arc *Arc[T], in BytesReader) (bool, error) {
	if err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING); err != nil {
		return false, err
	}

	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return false, err
	}
	return isBitSet(bitIndex, in)
}

// CountBits See BitTableUtil.countBits(int, Fst.BytesReader).
// The count of bit set is the number of arcs of a direct addressing node.
func CountBits[T any](arc *Arc[T], in BytesReader) (int64, error) {
	if err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING); err != nil {
		return 0, err
	}

	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}

	numPresenceBytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return 0, err
	}
	return countBits(numPresenceBytes, in)
}

// CountBitsUpTo See BitTableUtil.countBitsUpTo(int, Fst.BytesReader).
func CountBitsUpTo[T any](bitIndex int, arc *Arc[T], in BytesReader) (int, error) {
	if err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING); err != nil {
		return 0, err
	}

	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}
	return countBitsUpTo(bitIndex, in)
}

// NextBitSet See BitTableUtil.NextBitSet(int, int, Fst.BytesReader).
func NextBitSet[T any](bitIndex int, arc *Arc[T], in BytesReader) (int, error) {
	if err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING); err != nil {
		return 0, err
	}
	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}

	bytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return 0, err
	}
	return nextBitSet(bitIndex, int(bytes), in)
}

// PreviousBitSet See BitTableUtil.previousBitSet(int, Fst.BytesReader).
func PreviousBitSet[T any](bitIndex int, arc *Arc[T], in BytesReader) (int, error) {
	if err := assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING); err != nil {
		return 0, err
	}

	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}
	return previousBitSet(bitIndex, in)
}

// AssertIsValid Asserts the bit-table of the provided Fst.Arc is valid.
func AssertIsValid[T any](arc *Arc[T], in BytesReader) (bool, error) {
	err := assert(arc.BytesPerArc() > 0)
	if err != nil {
		return false, err
	}
	err = assert(arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return false, err
	}
	// First bit must be set.
	ok, err := IsBitSet(0, arc, in)
	if err != nil {
		return false, err
	}
	err = assert(ok)
	if err != nil {
		return false, err
	}
	// Last bit must be set.
	ok, err = IsBitSet(int(arc.NumArcs()-1), arc, in)
	if err != nil {
		return false, err
	}
	err = assert(ok)
	if err != nil {
		return false, err
	}
	// No bit set after the last arc.
	bitSet, err := NextBitSet(int(arc.NumArcs()-1), arc, in)
	if err != nil {
		return false, err
	}
	err = assert(bitSet == -1)
	if err != nil {
		return false, err
	}
	return true, nil
}
