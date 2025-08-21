package lucene80

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.NumericDocValues = &EmptyNumeric{}

type EmptyNumeric struct {
	doc int
}

func NewEmptyNumeric() *EmptyNumeric {
	return &EmptyNumeric{doc: -1}
}

func (e *EmptyNumeric) DocID() int {
	return e.doc
}

func (e *EmptyNumeric) NextDoc(ctx context.Context) (int, error) {
	e.doc = types.NO_MORE_DOCS
	return 0, io.EOF
}

func (e *EmptyNumeric) Advance(ctx context.Context, target int) (int, error) {
	e.doc = types.NO_MORE_DOCS
	return 0, io.EOF
}

func (e *EmptyNumeric) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, e, target)
}

func (e *EmptyNumeric) Cost() int64 {
	return 0
}

func (e *EmptyNumeric) AdvanceExact(target int) (bool, error) {
	e.doc = target
	return false, nil
}

func (e *EmptyNumeric) LongValue() (int64, error) {
	return 0, nil
}
