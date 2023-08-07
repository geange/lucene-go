package index

import "github.com/geange/lucene-go/core/types"

// NumericDocValues A per-document numeric item.
type NumericDocValues interface {
	types.DocValuesIterator

	// LongValue Returns the numeric item for the current document ID. It is illegal to call this method
	// after advanceExact(int) returned false.
	// Returns: numeric item
	LongValue() (int64, error)
}

var _ NumericDocValues = &NumericDocValuesDefault{}

type NumericDocValuesDefault struct {
	FnDocID        func() int
	FnNextDoc      func() (int, error)
	FnAdvance      func(target int) (int, error)
	FnSlowAdvance  func(target int) (int, error)
	FnCost         func() int64
	FnAdvanceExact func(target int) (bool, error)
	FnLongValue    func() (int64, error)
}

func (n *NumericDocValuesDefault) DocID() int {
	return n.FnDocID()
}

func (n *NumericDocValuesDefault) NextDoc() (int, error) {
	return n.FnNextDoc()
}

func (n *NumericDocValuesDefault) Advance(target int) (int, error) {
	return n.FnAdvance(target)
}

func (n *NumericDocValuesDefault) SlowAdvance(target int) (int, error) {
	return n.FnSlowAdvance(target)
}

func (n *NumericDocValuesDefault) Cost() int64 {
	return n.FnCost()
}

func (n *NumericDocValuesDefault) AdvanceExact(target int) (bool, error) {
	return n.FnAdvanceExact(target)
}

func (n *NumericDocValuesDefault) LongValue() (int64, error) {
	return n.FnLongValue()
}
