package sorter

import (
	"sort"
)

const (
	// LEVEL_THRESHOLD
	// after that many levels of recursion we fall back to introsort anyway
	// this is used as a protection against the fact that radix sort performs
	// worse when there are long common prefixes (probably because of cache
	// locality)
	LEVEL_THRESHOLD = 8

	// HISTOGRAM_SIZE
	// size of histograms: 256 + 1 to indicate that the string is finished
	HISTOGRAM_SIZE = 257

	// LENGTH_THRESHOLD
	// buckets below this size will be sorted with introsort
	LENGTH_THRESHOLD = 100
)

type MSBRadixInterface interface {
	Compare(i, j, skipBytes int) int
	Swap(i, j int)
	ByteAt(i int, k int) int
	//Value(i int) []byte
}

type msbRadixSorter struct {
	sorter MSBRadixInterface

	// we store one histogram per recursion level
	histograms   [LEVEL_THRESHOLD][]int
	endOffsets   []int
	commonPrefix []int
	maxLength    int
}

func NewMsbRadixSorter(maxLength int, sorter MSBRadixInterface) Sorter {
	return &msbRadixSorter{
		sorter:       sorter,
		commonPrefix: make([]int, min(24, maxLength)),
		maxLength:    maxLength,
		endOffsets:   make([]int, HISTOGRAM_SIZE),
	}
}

func (s *msbRadixSorter) Sort(from, to int) {
	s.sortKthAndLevel(from, to, 0, 0)
}

func (s *msbRadixSorter) sortKthAndLevel(from, to, kth, level int) {
	if to-from <= LENGTH_THRESHOLD || level >= LEVEL_THRESHOLD {
		s.pdqSort(from, to, kth)
	} else {
		s.radixSort(from, to, kth, level)
	}
}

func (s *msbRadixSorter) pdqSort(from, to, k int) {
	sort.Sort(&introSorterKPrefix{
		k:      k,
		from:   from,
		to:     to,
		sorter: s.sorter,
	})
}

var _ sort.Interface = &introSorterKPrefix{}

type introSorterKPrefix struct {
	k        int
	from, to int
	sorter   MSBRadixInterface
}

func (r *introSorterKPrefix) Len() int {
	return r.to - r.from
}

func (r *introSorterKPrefix) Less(i, j int) bool {
	//b1 := r.sorter.Value(r.from + i)[r.k:]
	//b2 := r.sorter.Value(r.from + j)[r.k:]
	//return bytes.Compare(b1, b2) < 0
	return r.sorter.Compare(r.from+i, r.from+j, r.k) < 0
}

func (r *introSorterKPrefix) Swap(i, j int) {
	r.sorter.Swap(r.from+i, r.from+j)
}

func (s *msbRadixSorter) getFallbackSorter(k int) sort.Interface {
	return &introSorterKPrefix{
		k:      k,
		sorter: s.sorter,
	}
}

func (s *msbRadixSorter) radixSort(from int, to int, k int, level int) {
	histogram := s.histograms[level]
	if len(histogram) == 0 {
		histogram = make([]int, HISTOGRAM_SIZE)
		s.histograms[level] = histogram
	} else {
		clear(histogram)
	}

	commonPrefixLength := s.prefixLengthAndBuildHistogram(from, to, k, histogram)
	if commonPrefixLength > 0 {
		// if there are no more chars to compare or if all entries fell into the
		// first bucket (which means strings are shorter than k) then we are done
		// otherwise recurse
		if k+commonPrefixLength < s.maxLength && histogram[0] < to-from {
			s.radixSort(from, to, k+commonPrefixLength, level)
		}
		return
	}

	startOffsets := histogram
	endOffsets := s.endOffsets
	s.sumHistogram(histogram, endOffsets)
	s.reorder(from, to, startOffsets, endOffsets, k)
	endOffsets = startOffsets

	if k+1 < s.maxLength {
		// recurse on all but the first bucket since all keys are equals in this
		// bucket (we already compared all bytes)
		for prev, i := endOffsets[0], 1; i < HISTOGRAM_SIZE; i++ {
			h := endOffsets[i]
			bucketLen := h - prev
			if bucketLen > 1 {
				s.sortKthAndLevel(from+prev, from+h, k+1, level+1)
			}
			prev = h
		}
	}
}

// Build a histogram of the number of values per bucket and return a common prefix length for all visited values.
// See Also: buildHistogram
// computeCommonPrefixLengthAndBuildHistogram is too long
func (s *msbRadixSorter) prefixLengthAndBuildHistogram(from, to, k int, histogram []int) (commonPrefixLength int) {
	commonPrefix := s.commonPrefix
	commonPrefixLength = min(len(commonPrefix), s.maxLength-k)

	for j := 0; j < commonPrefixLength; j++ {
		b := s.sorter.ByteAt(from, k+j)
		commonPrefix[j] = b
		if b == -1 {
			commonPrefixLength = j + 1
			break
		}
	}

	var i int

OUTER:
	for i = from + 1; i < to; i++ {
		for j := 0; j < commonPrefixLength; j++ {
			b := s.sorter.ByteAt(i, k+j)
			if b != commonPrefix[j] {
				commonPrefixLength = j
				if commonPrefixLength == 0 { // we have no common prefix
					histogram[int(commonPrefix[0])+1] = i - from
					histogram[b+1] = 1
					break OUTER
				}
				break
			}
		}
	}

	if i < to {
		// the loop got broken because there is no common prefix
		s.buildHistogram(i+1, to, k, histogram)
	} else {
		histogram[int(commonPrefix[0])+1] = to - from
	}

	return commonPrefixLength
}

// Reorder based on start/end offsets for each bucket.
// When this method returns, startOffsets and endOffsets are equal.
// startOffsets: start offsets per bucket
// endOffsets: end offsets per bucket
func (s *msbRadixSorter) reorder(from, to int, startOffsets, endOffsets []int, k int) {
	// reorder in place, like the dutch flag problem
	for i := 0; i < HISTOGRAM_SIZE; i++ {
		limit := endOffsets[i]
		for h1 := startOffsets[i]; h1 < limit; h1 = startOffsets[i] {
			// 获取k位定字节
			b := s.getBucket(from+h1, k)

			// 获取这个字节的桶的开始位置
			h2 := startOffsets[b]

			// 当前桶的开始位置+1
			startOffsets[b]++

			// 将 from + h1 放到 from + h2 的位置
			s.sorter.Swap(from+h1, from+h2)
		}
	}
}

// Build an histogram of the k-th characters of values occurring between offsets from and to,
// using getBucket.
func (s *msbRadixSorter) buildHistogram(from, to, k int, histogram []int) {
	for i := from; i < to; i++ {
		idx := s.getBucket(i, k)
		histogram[idx]++
	}
}

// Accumulate values of the histogram so that it does not store counts but start offsets.
// endOffsets will store the end offsets.
func (s *msbRadixSorter) sumHistogram(histogram []int, endOffsets []int) {
	accum := 0
	for i := 0; i < HISTOGRAM_SIZE; i++ {
		count := histogram[i]
		histogram[i] = accum
		accum += count
		endOffsets[i] = accum
	}
}

// Return a number for the k-th character between 0 and HISTOGRAM_SIZE.
func (s *msbRadixSorter) getBucket(i int, k int) int {
	return s.sorter.ByteAt(i, k) + 1
}
