package fst

// InputOutput Holds a single input (BytesRef) + output pair.
type InputOutput[T PairAble] struct {
	Input  []byte
	Output T
}
