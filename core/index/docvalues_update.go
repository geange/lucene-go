package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
)

type DocValuesUpdate interface {
	ValueSizeInBytes() int64
	ValueToString() string
	WriteTo(output store.DataOutput) error
	HasValue() bool

	GetOptions() *DocValuesUpdateOptions
}

type DocValuesUpdateOptions struct {
	DType     document.DocValuesType
	Term      *Term
	Field     string
	DocIDUpto int
	HasValue  bool
}

var _ DocValuesUpdate = &NumericDocValuesUpdate{}

type NumericDocValuesUpdate struct {
}

func NewNumericDocValuesUpdate(term *Term, field string, value int64) *NumericDocValuesUpdate {
	panic("")
}

func (n *NumericDocValuesUpdate) GetOptions() *DocValuesUpdateOptions {
	//TODO implement me
	panic("implement me")
}

func (n *NumericDocValuesUpdate) ValueSizeInBytes() int64 {
	//TODO implement me
	panic("implement me")
}

func (n *NumericDocValuesUpdate) ValueToString() string {
	//TODO implement me
	panic("implement me")
}

func (n *NumericDocValuesUpdate) WriteTo(output store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (n *NumericDocValuesUpdate) HasValue() bool {
	//TODO implement me
	panic("implement me")
}

func (n *NumericDocValuesUpdate) GetDocValuesType() document.DocValuesType {
	//TODO implement me
	panic("implement me")
}

var _ DocValuesUpdate = &BinaryDocValuesUpdate{}

type BinaryDocValuesUpdate struct {
}

func NewBinaryDocValuesUpdate(term *Term, field string, value []byte) *BinaryDocValuesUpdate {
	panic("")
}

func (b *BinaryDocValuesUpdate) GetOptions() *DocValuesUpdateOptions {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesUpdate) ValueSizeInBytes() int64 {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesUpdate) ValueToString() string {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesUpdate) WriteTo(output store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesUpdate) HasValue() bool {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesUpdate) GetDocValuesType() document.DocValuesType {
	//TODO implement me
	panic("implement me")
}
