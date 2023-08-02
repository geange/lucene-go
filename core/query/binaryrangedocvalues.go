package query

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.BinaryDocValues = &BinaryRangeDocValues{}

type BinaryRangeDocValues struct {
	in                   index.BinaryDocValues
	packedValue          []byte
	numDims              int
	numBytesPerDimension int
	docID                int
}

func NewBinaryRangeDocValues(in index.BinaryDocValues, numDims int, numBytesPerDimension int) *BinaryRangeDocValues {
	return &BinaryRangeDocValues{
		in:                   in,
		numDims:              numDims,
		numBytesPerDimension: numBytesPerDimension,
		packedValue:          make([]byte, 2*numDims*numBytesPerDimension),
	}
}

func (b *BinaryRangeDocValues) DocID() int {
	return b.in.DocID()
}

func (b *BinaryRangeDocValues) NextDoc() (int, error) {
	docID, err := b.in.NextDoc()
	if err != nil {
		return 0, err
	}
	if err := b.decodeRanges(); err != nil {
		return 0, err
	}
	return docID, nil
}

func (b *BinaryRangeDocValues) Advance(target int) (int, error) {
	res, err := b.in.Advance(target)
	if err != nil {
		return 0, err
	}

	if err := b.decodeRanges(); err != nil {
		return 0, err
	}

	return res, nil
}

func (b *BinaryRangeDocValues) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(b, target)
}

func (b *BinaryRangeDocValues) Cost() int64 {
	return b.in.Cost()
}

func (b *BinaryRangeDocValues) AdvanceExact(target int) (bool, error) {
	res, err := b.in.AdvanceExact(target)
	if err != nil {
		return false, err
	}
	if res {
		if err := b.decodeRanges(); err != nil {
			return false, err
		}
	}
	return res, nil
}

func (b *BinaryRangeDocValues) BinaryValue() ([]byte, error) {
	return b.in.BinaryValue()
}

func (b *BinaryRangeDocValues) getPackedValue() []byte {
	return b.packedValue
}

func (b *BinaryRangeDocValues) decodeRanges() error {
	binaryValue, err := b.in.BinaryValue()
	if err != nil {
		return err
	}
	// We reuse the existing allocated memory for packed values since all docvalues in this iterator
	// should be exactly same in indexed structure, hence the byte representations in length should be identical
	copy(b.packedValue, binaryValue[:2*b.numDims*b.numBytesPerDimension])
	return nil
}
