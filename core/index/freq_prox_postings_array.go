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
