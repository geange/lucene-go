package index

var _ TermState = &OrdTermState{}

type OrdTermState struct {
	Ord int64
}

func NewOrdTermState() *OrdTermState {
	return &OrdTermState{}
}

func (r *OrdTermState) CopyFrom(other TermState) {
	if v, ok := other.(*OrdTermState); ok {
		r.Ord = v.Ord
	}
}
