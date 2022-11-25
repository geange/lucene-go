package fst

// EnumBase doFloor controls the behavior of advance: if it's true doFloor is true,
// advance positions to the biggest term before target.
type EnumBase struct {
	fst          *FST
	arcs         []*Arc
	output       []any
	fstReader    BytesReader
	upto         int
	targetLength int

	spi EnumBaseSPI
}

type EnumBaseSPI interface {
	GetTargetLabel() (int, error)
	GetCurrentLabel() (int, error)
	SetCurrentLabel(label int) error
	Grow() error
}

// Rewinds enum state to match the shared prefix between current term and target term
func (e *EnumBase) rewindPrefix() error {
	if e.upto == 0 {
		//System.out.println("  init");
		e.upto = 1
		//e.fst.ReadFirstTargetArc(getArc(0), getArc(1), fstReader)
		//return
	}

	return nil
}
