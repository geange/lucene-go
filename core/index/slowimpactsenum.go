package index

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"math"
)

var _ index.ImpactsEnum = &SlowImpactsEnum{}

type SlowImpactsEnum struct {
	delegate index.PostingsEnum
}

func NewSlowImpactsEnum(delegate index.PostingsEnum) *SlowImpactsEnum {
	return &SlowImpactsEnum{delegate: delegate}
}

func (s *SlowImpactsEnum) DocID() int {
	return s.delegate.DocID()
}

func (s *SlowImpactsEnum) NextDoc() (int, error) {
	return s.delegate.NextDoc()
}

func (s *SlowImpactsEnum) Advance(target int) (int, error) {
	return s.delegate.Advance(target)
}

func (s *SlowImpactsEnum) SlowAdvance(target int) (int, error) {
	return s.Advance(target)
}

func (s *SlowImpactsEnum) Cost() int64 {
	return s.delegate.Cost()
}

func (s *SlowImpactsEnum) Freq() (int, error) {
	return s.delegate.Freq()
}

func (s *SlowImpactsEnum) NextPosition() (int, error) {
	return s.delegate.NextPosition()
}

func (s *SlowImpactsEnum) StartOffset() (int, error) {
	return s.delegate.StartOffset()
}

func (s *SlowImpactsEnum) EndOffset() (int, error) {
	return s.delegate.EndOffset()
}

func (s *SlowImpactsEnum) GetPayload() ([]byte, error) {
	return s.delegate.GetPayload()
}

func (s *SlowImpactsEnum) AdvanceShallow(target int) error {
	return nil
}

func (s *SlowImpactsEnum) GetImpacts() (index.Impacts, error) {
	return dummyImpacts, nil
}

var _ index.Impacts = &slowImpactsEnumImpacts{}

type slowImpactsEnumImpacts struct {
	impacts []index.Impact
}

func (s *slowImpactsEnumImpacts) NumLevels() int {
	return 1
}

var dummyImpacts = &slowImpactsEnumImpacts{
	impacts: []index.Impact{NewImpact(math.MaxInt32, 1)},
}

func (s *slowImpactsEnumImpacts) GetDocIdUpTo(level int) int {
	return types.NO_MORE_DOCS
}

func (s *slowImpactsEnumImpacts) GetImpacts(level int) []index.Impact {
	return s.impacts
}
