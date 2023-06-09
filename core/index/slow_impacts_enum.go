package index

import "math"

var _ ImpactsEnum = &SlowImpactsEnum{}

type SlowImpactsEnum struct {
	delegate PostingsEnum
}

func NewSlowImpactsEnum(delegate PostingsEnum) *SlowImpactsEnum {
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

func (s *SlowImpactsEnum) GetImpacts() (Impacts, error) {
	return dummyImpacts, nil
}

var _ Impacts = &slowImpactsEnumImpacts{}

type slowImpactsEnumImpacts struct {
	impacts []*Impact
}

func (s *slowImpactsEnumImpacts) NumLevels() int {
	return 1
}

var dummyImpacts = &slowImpactsEnumImpacts{
	impacts: []*Impact{NewImpact(math.MaxInt32, 1)},
}

func (s *slowImpactsEnumImpacts) GetDocIdUpTo(level int) int {
	return NO_MORE_DOCS
}

func (s *slowImpactsEnumImpacts) GetImpacts(level int) []*Impact {
	return s.impacts
}
