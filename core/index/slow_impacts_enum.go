package index

var _ ImpactsEnum = &SlowImpactsEnum{}

type SlowImpactsEnum struct {
	delegate PostingsEnum
}

func NewSlowImpactsEnum(delegate PostingsEnum) *SlowImpactsEnum {
	return &SlowImpactsEnum{delegate: delegate}
}

func (s *SlowImpactsEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) GetPayload() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) AdvanceShallow(target int) error {
	//TODO implement me
	panic("implement me")
}

func (s *SlowImpactsEnum) GetImpacts() (Impacts, error) {
	//TODO implement me
	panic("implement me")
}
