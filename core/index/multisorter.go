package index

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util"
)

// SortCodecReader
// Does a merge sort of the leaves of the incoming reader, returning MergeState.DocMap to map each leaf's documents into the merged segment. The documents for each incoming leaf reader must already be sorted by the same sort! Returns null if the merge sort is not needed (segments are already in index sort order).
func SortCodecReader(sort index.Sort, readers []index.CodecReader) ([]MergeStateDocMap, error) {
	//fields := sort.GetSort()
	//
	//comparables := make([][]ComparableProvider, len(fields))
	//reverseMuls := make([]int, len(fields))
	//for _, field := range fields {
	//	sorter := field.GetIndexSorter()
	//	if sorter == nil {
	//		return nil, fmt.Errorf("cannot use sort field:%s for index sorting", field)
	//	}
	//	comparables[i] = sorter.get
	//}

	// TODO
	panic("")
}

type LeafAndDocID struct {
	readerIndex             int
	liveDocs                util.Bits
	maxDoc                  int
	valuesAsComparableLongs []int64
	docId                   int
}

func NewLeafAndDocID(readerIndex int, liveDocs util.Bits, maxDoc int, numComparables int) *LeafAndDocID {
	return &LeafAndDocID{
		readerIndex:             readerIndex,
		liveDocs:                liveDocs,
		maxDoc:                  maxDoc,
		valuesAsComparableLongs: make([]int64, numComparables),
	}
}
