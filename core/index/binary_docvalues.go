package index

// BinaryDocValues A per-document numeric value.
type BinaryDocValues interface {
	DocValuesIterator

	// BinaryValue Returns the binary value for the current document ID. It is illegal to call this method after
	// advanceExact(int) returned false.
	// Returns: binary value
	BinaryValue() ([]byte, error)
}

type BinaryDocValuesDefault struct {
	FnDocID        func() int
	FnNextDoc      func() (int, error)
	FnAdvance      func(target int) (int, error)
	FnSlowAdvance  func(target int) (int, error)
	FnCost         func() int64
	FnAdvanceExact func(target int) (bool, error)
	FnBinaryValue  func() ([]byte, error)
}

func (n *BinaryDocValuesDefault) DocID() int {
	return n.FnDocID()
}

func (n *BinaryDocValuesDefault) NextDoc() (int, error) {
	return n.FnNextDoc()
}

func (n *BinaryDocValuesDefault) Advance(target int) (int, error) {
	return n.FnAdvance(target)
}

func (n *BinaryDocValuesDefault) SlowAdvance(target int) (int, error) {
	return n.FnSlowAdvance(target)
}

func (n *BinaryDocValuesDefault) Cost() int64 {
	return n.FnCost()
}

func (n *BinaryDocValuesDefault) AdvanceExact(target int) (bool, error) {
	return n.FnAdvanceExact(target)
}

func (n *BinaryDocValuesDefault) BinaryValue() ([]byte, error) {
	return n.FnBinaryValue()
}
