package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"math"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
)

// DocValuesUpdate
// An in-place update to a DocValues field.
type DocValuesUpdate interface {
	GetType() document.DocValuesType
	GetTerm() index.Term
	GetField() string
	GetDocIDUpto() int
	GetHasValue() bool
	WriteTo(ctx context.Context, out store.DataOutput) error
	HasValue() bool
}

type DocValuesUpdateOptions struct {
	DType     document.DocValuesType
	Term      index.Term
	Field     string
	DocIDUpto int
	HasValue  bool
}

var _ DocValuesUpdate = &NumericDocValuesUpdate{}

type BaseDocValuesUpdate struct {
	_type     document.DocValuesType
	term      index.Term
	field     string
	docIDUpto int
	hasValue  bool
}

func (d *BaseDocValuesUpdate) GetType() document.DocValuesType {
	return d._type
}

func (d *BaseDocValuesUpdate) GetTerm() index.Term {
	return d.term
}

func (d *BaseDocValuesUpdate) GetField() string {
	return d.field
}

func (d *BaseDocValuesUpdate) GetDocIDUpto() int {
	return d.docIDUpto
}

func (d *BaseDocValuesUpdate) GetHasValue() bool {
	return d.hasValue
}

type NumericDocValuesUpdate struct {
	BaseDocValuesUpdate

	value int64
}

func NewNumericDocValuesUpdate(term index.Term, field string, value int64) *NumericDocValuesUpdate {
	return newNumericDocValuesUpdate(term, field, value, math.MaxInt32, true)
}

func newNumericDocValuesUpdate(term index.Term, field string, value int64, docIDUpTo int, hasValue bool) *NumericDocValuesUpdate {
	return &NumericDocValuesUpdate{
		BaseDocValuesUpdate: BaseDocValuesUpdate{
			_type:     document.DOC_VALUES_TYPE_NUMERIC,
			term:      term,
			field:     field,
			docIDUpto: docIDUpTo,
			hasValue:  hasValue,
		},
		value: value,
	}
}

func (n *NumericDocValuesUpdate) WriteTo(ctx context.Context, out store.DataOutput) error {
	return out.WriteUvarint(ctx, uint64(n.value))
}

func (n *NumericDocValuesUpdate) HasValue() bool {
	return n.hasValue
}

func (n *NumericDocValuesUpdate) GetValue() int64 {
	return n.value
}

var _ DocValuesUpdate = &BinaryDocValuesUpdate{}

type BinaryDocValuesUpdate struct {
	BaseDocValuesUpdate

	value []byte
}

func NewBinaryDocValuesUpdate(term index.Term, field string, value []byte) *BinaryDocValuesUpdate {
	return newBinaryDocValuesUpdate(term, field, value, math.MaxInt32)
}

func newBinaryDocValuesUpdate(term index.Term, field string, value []byte, docIDUpTo int) *BinaryDocValuesUpdate {
	return &BinaryDocValuesUpdate{
		BaseDocValuesUpdate: BaseDocValuesUpdate{
			_type:     document.DOC_VALUES_TYPE_BINARY,
			term:      term,
			field:     field,
			docIDUpto: docIDUpTo,
			hasValue:  len(value) == 0,
		},
		value: value,
	}
}

func (b *BinaryDocValuesUpdate) WriteTo(ctx context.Context, out store.DataOutput) error {
	err := out.WriteUvarint(ctx, uint64(len(b.value)))
	if err != nil {
		return err
	}
	_, err = out.Write(b.value)
	return err
}

func (b *BinaryDocValuesUpdate) HasValue() bool {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesUpdate) GetValue() []byte {
	return b.value
}
