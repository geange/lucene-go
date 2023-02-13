package util

import "bytes"

// RadixSelector Radix selector.
// This implementation works similarly to a MSB radix sort except that it only recurses
// into the sub partition that contains the desired value.
// lucene.internal
//type RadixSelector interface {
//	Selector
//
//	// ByteAt Return the k-th byte of the entry at index i, or -1 if its length is less than or
//	// equal to k. This may only be called with a value of i between 0 included and maxLength excluded.
//	ByteAt(i, k int) int
//}

const (

	// LEVEL_THRESHOLD after that many levels of recursion we fall back to introselect anyway
	// this is used as a protection against the fact that radix sort performs
	// worse when there are long common prefixes (probably because of cache
	// locality)
	LEVEL_THRESHOLD = 8

	// HISTOGRAM_SIZE size of histograms: 256 + 1 to indicate that the string is finished
	HISTOGRAM_SIZE = 257

	// LENGTH_THRESHOLD buckets below this size will be sorted with introselect
	LENGTH_THRESHOLD = 100
)

var _ Selector = &RadixSelector{}

type RadixSelectorConfig struct {
	MaxLength             int
	FnByteAt              func(i, k int) int
	FnSwap                func(i, j int)
	FnGetFallbackSelector func(d int) Selector
}

type RadixSelector struct {
	// we store one histogram per recursion level
	histogram             []int
	commonPrefix          []int
	maxLength             int
	fnByteAt              func(i, k int) int
	fnSwap                func(i, j int)
	fnGetFallbackSelector func(d int) Selector
}

// NewRadixSelector Sole constructor.
// Params: maxLength – the maximum length of keys, pass Integer.MAX_VALUE if unknown.
func NewRadixSelector(cfg *RadixSelectorConfig) *RadixSelector {
	return &RadixSelector{
		histogram:    make([]int, HISTOGRAM_SIZE),
		commonPrefix: make([]int, Min(24, cfg.MaxLength)),
		maxLength:    cfg.MaxLength,
		fnByteAt:     cfg.FnByteAt,
		fnSwap:       cfg.FnSwap,
	}
}

func (r *RadixSelector) Swap(i, j int) {
	r.fnSwap(i, j)
}

func (r *RadixSelector) getFallbackSelector(d int) Selector {
	buf := new(bytes.Buffer)

	return NewIntroSelector(&IntroSelectorConfig{
		FnSwap: func(i, j int) {
			r.fnSwap(i, j)
		},
		FnSetPivot: func(i int) {
			buf.Reset()
			for o := d; o < r.maxLength; o++ {
				b := r.fnByteAt(i, o)
				if b == -1 {
					break
				}
				buf.WriteByte(byte(b))
			}
		},
		FnComparePivot: func(j int) int {
			bs := buf.Bytes()

			for i, b1 := range bs {
				b2 := r.fnByteAt(j, d+i)
				if int(b1) != b2 {
					return int(b1) - b2
				}
			}

			if d+buf.Len() == r.maxLength {
				return 0
			}
			return -1 - r.fnByteAt(j, d+buf.Len())
		},
		FnCompare: func(i, j int) int {
			for o := d; o < r.maxLength; o++ {
				b1 := r.fnByteAt(i, o)
				b2 := r.fnByteAt(j, o)
				if b1 != b2 {
					return b1 - b2
				} else if b1 == -1 {
					break
				}
			}
			return 0
		},
	})
}

func (r *RadixSelector) Select(from, to, k int) {
	SelectorCheckArgs(from, to, k)
	r.fnSelect(from, to, k, 0, 0)
}

func (r *RadixSelector) fnSelect(from, to, k, d, l int) {
	if to-from <= LENGTH_THRESHOLD || l >= LEVEL_THRESHOLD {
		r.getFallbackSelector(d).Select(from, to, k)
	} else {
		r.radixSelect(from, to, k, d, l)
	}
}

// Params:  d – the character number to Compare
//
//	l – the level of recursion
func (r *RadixSelector) radixSelect(from, to, k, d, l int) {
	histogram := r.histogram
	for i := range histogram {
		histogram[i] = 0
	}

	commonPrefixLength := r.computeCommonPrefixLengthAndBuildHistogram(from, to, d, histogram)
	if commonPrefixLength > 0 {
		// if there are no more chars to Compare or if all entries fell into the
		// first bucket (which means strings are shorter than d) then we are done
		// otherwise recurse
		if d+commonPrefixLength < r.maxLength && histogram[0] < to-from {
			r.radixSelect(from, to, k, d+commonPrefixLength, l)
		}
		return
	}

	if !r.assertHistogram(commonPrefixLength, histogram) {
		panic("")
	}

	bucketFrom := from
	for bucket := 0; bucket < HISTOGRAM_SIZE; bucket++ {
		bucketTo := bucketFrom + histogram[bucket]

		if bucketTo > k {
			r.partition(from, to, bucket, bucketFrom, bucketTo, d)

			if bucket != 0 && d+1 < r.maxLength {
				// all elements in bucket 0 are equal so we only need to recurse if bucket != 0
				r.fnSelect(bucketFrom, bucketTo, k, d+1, l+1)
			}
			return
		}
		bucketFrom = bucketTo
	}
}

// only used from assert
func (r *RadixSelector) assertHistogram(commonPrefixLength int, histogram []int) bool {
	numberOfUniqueBytes := 0
	for _, freq := range r.histogram {
		if freq > 0 {
			numberOfUniqueBytes++
		}
	}
	if numberOfUniqueBytes == 1 {
		if !(commonPrefixLength >= 1) {
			panic("")
		}
	} else {
		if !(commonPrefixLength == 0) {
			panic("")
		}
	}
	return true
}

// Return a number for the k-th character between 0 and HISTOGRAM_SIZE.
func (r *RadixSelector) getBucket(i, k int) int {
	return r.fnByteAt(i, k) + 1
}

// Build a histogram of the number of values per bucket and return a common prefix length for all visited values.
// See Also: buildHistogram
func (r *RadixSelector) computeCommonPrefixLengthAndBuildHistogram(from, to, k int, histogram []int) int {
	panic("")
}

// Build an histogram of the k-th characters of values occurring between offsets from and to, using getBucket.
func (r *RadixSelector) buildHistogram(from, to, k int, histogram []int) {
	for i := from; i < to; i++ {
		histogram[r.getBucket(i, k)]++
	}
}

// Reorder elements so that all of them that fall into bucket are between offsets bucketFrom and bucketTo.
func (r *RadixSelector) partition(from, to, bucket, bucketFrom, bucketTo, d int) {
	left := from
	right := to - 1

	slot := bucketFrom

	for {
		leftBucket := r.getBucket(left, d)
		rightBucket := r.getBucket(right, d)

		for leftBucket <= bucket && left < bucketFrom {
			if leftBucket == bucket {
				r.fnSwap(left, slot)
				slot++
			} else {
				left++
			}
			leftBucket = r.getBucket(left, d)
		}

		for rightBucket >= bucket && right >= bucketTo {
			if rightBucket == bucket {
				r.fnSwap(right, slot)
				slot++
			} else {
				right--
			}
			rightBucket = r.getBucket(right, d)
		}

		if left < bucketFrom && right >= bucketTo {
			r.fnSwap(left, right)
			left++
			right--
		} else {
			//assert left == bucketFrom;
			//assert right == bucketTo - 1;
			break
		}
	}
}
