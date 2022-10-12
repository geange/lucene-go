package fst

func IsBitSet[T any](bitIndex int, arc *Arc[T], in BytesReader) bool {
	in.SetPosition(arc.bitTableStart)
	return BitTableUtil.isBitSet(bitIndex, in)
}

func NextBitSet[T any](bitIndex int, arc *Arc[T], in BytesReader) int {
	in.SetPosition(arc.bitTableStart)
	return BitTableUtil.nextBitSet(bitIndex, getNumPresenceBytes(arc.NumArcs()), in)
}

func PreviousBitSet[T any](bitIndex int, arc *Arc[T], in BytesReader) int {
	in.SetPosition(arc.bitTableStart)
	return BitTableUtil.previousBitSet(bitIndex, in)
}

func CountBitsUpTo[T any](bitIndex int, arc *Arc[T], in BytesReader) int {
	in.SetPosition(arc.bitTableStart)
	return BitTableUtil.countBitsUpTo(bitIndex, in)
}

func CountBits[T any](arc *Arc[T], in BytesReader) int {
	in.SetPosition(arc.bitTableStart)
	return BitTableUtil.countBits(getNumPresenceBytes(arc.NumArcs()), in)
}
