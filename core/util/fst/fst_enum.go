package fst

//	doFloor controls the behavior of advance: if it's true doFloor is true,
//
// advance positions to the biggest term before target.
type Enum[T PairAble] struct {
	fst          *FST[T]
	arcs         []*Arc[T]
	output       []T
	noOutput     T
	fstReader    BytesReader
	upto         int
	targetLength int

	spi EnumSPI
}

func NewEnum[T PairAble](fst *FST[T]) (*Enum[T], error) {
	reader, err := fst.GetBytesReader()
	if err != nil {
		return nil, err
	}

	noOutput := fst.outputs.GetNoOutput()

	enum := &Enum[T]{
		fst:       fst,
		fstReader: reader,
		noOutput:  noOutput,
		output:    []T{noOutput},
	}

	if _, err := fst.GetFirstArc(enum.getArc(0)); err != nil {
		return nil, err
	}

	return enum, nil
}

type EnumSPI interface {
	GetTargetLabel() (int, error)
	GetCurrentLabel() (int, error)
	SetCurrentLabel(label int) error
	Grow() error
}

// Rewinds enum state to match the shared prefix between current term and target term
func (e *Enum[T]) rewindPrefix() error {
	if e.upto == 0 {
		//System.out.println("  init");
		e.upto = 1
		//e.fst.ReadFirstTargetArc(getArc(0), getArc(1), fstReader)
		//return
	}

	return nil
}

func (e *Enum[T]) getArc(idx int) *Arc[T] {
	if e.arcs[idx] == nil {
		e.arcs[idx] = new(Arc[T])
	}
	return e.arcs[idx]
}
