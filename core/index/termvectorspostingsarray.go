package index

var _ ParallelPostingsArray = &TermVectorsPostingsArray{}

type TermVectorsPostingsArray struct {
	*ParallelPostingsArrayDefault

	freqs         []int // How many times this term occurred in the current doc
	lastOffsets   []int // Last offset we saw
	lastPositions []int // Last position where this term occurred
}

func NewTermVectorsPostingsArray() *TermVectorsPostingsArray {
	return &TermVectorsPostingsArray{
		ParallelPostingsArrayDefault: NewParallelPostingsArrayDefault(),
		freqs:                        []int{},
		lastOffsets:                  []int{},
		lastPositions:                []int{},
	}
}

func (t *TermVectorsPostingsArray) NewInstance() ParallelPostingsArray {
	return &TermVectorsPostingsArray{
		ParallelPostingsArrayDefault: NewParallelPostingsArrayDefault(),
		freqs:                        []int{},
		lastOffsets:                  []int{},
		lastPositions:                []int{},
	}
}

func (t *TermVectorsPostingsArray) BytesPerPosting() int {
	return BYTES_PER_POSTING + 3*4
}

func (t *TermVectorsPostingsArray) Grow() {
	t.ParallelPostingsArrayDefault.Grow()
	t.freqs = append(t.freqs, 0)
	t.lastOffsets = append(t.lastOffsets, 0)
	t.lastPositions = append(t.lastPositions, 0)
}

func (t *TermVectorsPostingsArray) SetFreqs(termID, v int) {
	if termID >= len(t.freqs) {
		size := termID - len(t.freqs) + 1
		t.freqs = append(t.freqs, make([]int, size)...)
	}
	t.freqs[termID] = v
}

func (t *TermVectorsPostingsArray) SetLastOffsets(termID, v int) {
	if termID >= len(t.lastOffsets) {
		size := termID - len(t.lastOffsets) + 1
		t.lastOffsets = append(t.lastOffsets, make([]int, size)...)
	}
	t.lastOffsets[termID] = v
}

func (t *TermVectorsPostingsArray) SetLastPositions(termID, v int) {
	if termID >= len(t.lastPositions) {
		size := termID - len(t.lastPositions) + 1
		t.lastPositions = append(t.lastPositions, make([]int, size)...)
	}
	t.lastPositions[termID] = v
}
