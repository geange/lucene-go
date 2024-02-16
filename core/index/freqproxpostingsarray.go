package index

var _ ParallelPostingsArray = &FreqProxPostingsArray{}

type FreqProxPostingsArray struct {
	*ParallelPostingsArrayDefault

	termFreqs     []int // # times this term occurs in the current doc
	lastDocIDs    []int // Last docID where this term occurred
	lastDocCodes  []int // Code for prior doc
	lastPositions []int // Last position where this term occurred
	lastOffsets   []int // Last endOffset where this term occurred
}

func NewFreqProxPostingsArray(writeFreqs, writeProx, writeOffsets bool) *FreqProxPostingsArray {
	return &FreqProxPostingsArray{
		ParallelPostingsArrayDefault: NewParallelPostingsArrayDefault(),
		termFreqs:                    []int{},
		lastDocIDs:                   []int{},
		lastDocCodes:                 []int{},
		lastPositions:                []int{},
		lastOffsets:                  []int{},
	}
}

func (f *FreqProxPostingsArray) NewInstance() ParallelPostingsArray {
	return &FreqProxPostingsArray{
		ParallelPostingsArrayDefault: NewParallelPostingsArrayDefault(),
		termFreqs:                    []int{},
		lastDocIDs:                   []int{},
		lastDocCodes:                 []int{},
		lastPositions:                []int{},
		lastOffsets:                  []int{},
	}
}

func (f *FreqProxPostingsArray) SetTermFreqs(termID, v int) {
	if termID >= len(f.termFreqs) {
		size := termID - len(f.termFreqs) + 1
		f.termFreqs = append(f.termFreqs, make([]int, size)...)
	}
	f.termFreqs[termID] = v
}

func (f *FreqProxPostingsArray) SetLastDocIDs(termID, v int) {
	if termID >= len(f.lastDocIDs) {
		size := termID - len(f.lastDocIDs) + 1
		f.lastDocIDs = append(f.lastDocIDs, make([]int, size)...)
	}
	f.lastDocIDs[termID] = v
}

func (f *FreqProxPostingsArray) SetLastDocCodes(termID, v int) {
	if termID >= len(f.lastDocCodes) {
		size := termID - len(f.lastDocCodes) + 1
		f.lastDocCodes = append(f.lastDocCodes, make([]int, size)...)
	}
	f.lastDocCodes[termID] = v
}

func (f *FreqProxPostingsArray) SetLastPositions(termID, v int) {
	if termID >= len(f.lastPositions) {
		size := termID - len(f.lastPositions) + 1
		f.lastPositions = append(f.lastPositions, make([]int, size)...)
	}
	f.lastPositions[termID] = v
}

func (f *FreqProxPostingsArray) SetLastOffsets(termID, v int) {
	if termID >= len(f.lastOffsets) {
		size := termID - len(f.lastOffsets) + 1
		f.lastOffsets = append(f.lastOffsets, make([]int, size)...)
	}
	f.lastOffsets[termID] = v
}

func (f *FreqProxPostingsArray) BytesPerPosting() int {
	bytes := BYTES_PER_POSTING + 2*4
	if f.lastPositions != nil {
		bytes += 4
	}

	if f.lastOffsets != nil {
		bytes += 4
	}

	if f.termFreqs != nil {
		bytes += 4
	}

	return bytes
}
