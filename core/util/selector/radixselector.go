package selector

type RadixSelector interface {
	Swap(i, j int)
	ByteAt(i, k int) int
}

const (
	// LEVEL_THRESHOLD
	// 在经过许多级别的递归之后，我们无论如何都会回到introselect，
	// 这是为了防止基数排序在有长的公共前缀时表现更差（可能是因为缓存位置）
	LEVEL_THRESHOLD = 8

	HISTOGRAM_SIZE = 257

	LENGTH_THRESHOLD = 100
)

type radixSelector struct {
	radix        RadixSelector
	histogram    []int
	commonPrefix []int
	maxLength    int
}

func NewRadixSelector(radix RadixSelector, maxLength int) Selector {
	return &radixSelector{
		radix:        radix,
		maxLength:    maxLength,
		histogram:    make([]int, HISTOGRAM_SIZE),
		commonPrefix: make([]int, min(24, maxLength)),
	}
}

func (s *radixSelector) SelectK(from, to, k int) {
	s.selectK(from, to, k, 0, 0)
}

func (s *radixSelector) selectK(from, to, k, d, l int) {
	if to-from <= LENGTH_THRESHOLD || l >= LEVEL_THRESHOLD {
		s.getFallbackSelector(d).SelectK(from, to, k)
	} else {
		s.radixSelect(from, to, k, d, l)
	}
}

func (s *radixSelector) radixSelect(from, to, k, idx, level int) {
	histogram := s.histogram
	for i := range histogram {
		histogram[i] = 0
	}

	commonPrefixLength := s.computeCommonPrefixLengthAndBuildHistogram(from, to, idx, histogram)
	if commonPrefixLength > 0 {
		// if there are no more chars to compare or if all entries fell into the
		// first bucket (which means strings are shorter than d) then we are done
		// otherwise recurse
		// d+commonPrefixLength < s.maxLength 存在公共前缀，且公共前缀小于 maxLength
		// histogram[0] < to-from 非所有的slot都落到第一个桶里面，
		if idx+commonPrefixLength < s.maxLength && histogram[0] < to-from {
			s.radixSelect(from, to, k, idx+commonPrefixLength, level)
		}
		return
	}

	bucketFrom := from
	for bucket := 0; bucket < HISTOGRAM_SIZE; bucket++ {
		bucketTo := bucketFrom + histogram[bucket]

		if bucketTo > k {
			// 对bucket前后的桶进行分区处理
			// 重新排列元素的顺序，使所有落入bucket的元素都位于偏移bucketFrom和bucketTo之间。
			s.partition(from, to, bucket, bucketFrom, bucketTo, idx)

			// 对bucket所在的桶进行排序
			if bucket != 0 && idx+1 < s.maxLength {
				// all elements in bucket 0 are equal so we only need to recurse if bucket != 0
				s.selectK(bucketFrom, bucketTo, k, idx+1, level+1)
			}
			return
		}
		bucketFrom = bucketTo
	}
}

// Build a histogram of the number of values per bucket and return a common prefix length for all visited values.
func (s *radixSelector) computeCommonPrefixLengthAndBuildHistogram(from, to, idx int, histogram []int) int {
	commonPrefix := s.commonPrefix
	commonPrefixLength := min(len(commonPrefix), s.maxLength-idx)
	// 计算通用前缀可能的最大值
	// 把idx=from的值写入到commonPrefix
	for j := 0; j < commonPrefixLength; j++ {
		b := s.radix.ByteAt(from, idx+j)
		commonPrefix[j] = b
		if b == -1 {
			commonPrefixLength = j + 1
			break
		}

	}

	i := 0
OUTER:
	for i = from + 1; i < to; i++ {
		// 计算剩余的内容是否有公共前缀
		// 存在公共前缀的在 histogram[commonPrefix[0]+1] 记录具有公共前缀的数量
		// 如果部分存在公共前缀，则，记录这部分的数量在 histogram[commonPrefix[0]+1] 中
		for j := 0; j < commonPrefixLength; j++ {
			b := s.radix.ByteAt(i, idx+j)
			if b != commonPrefix[j] {
				commonPrefixLength = j
				if commonPrefixLength == 0 {
					// we have no common prefix
					histogram[commonPrefix[0]+1] = i - from // from到i的都有前缀
					histogram[b+1] = 1
					break OUTER
				}
				break
			}
		}
	}

	if i < to {
		// 剩余不具备公共前缀的，按照idx计算桶元素的数量
		s.buildHistogram(i+1, to, idx, histogram)
	} else {
		// 全部相同，则记录全部的数量到一个桶内
		histogram[commonPrefix[0]+1] = to - from
	}
	return commonPrefixLength
}

// Reorder elements so that all of them that fall into bucket are between offsets bucketFrom and bucketTo.
func (s *radixSelector) partition(from, to, bucket, bucketFrom, bucketTo, d int) {
	left := from
	right := to - 1
	slot := bucketFrom

	for {
		leftBucket := s.getBucket(left, d)
		rightBucket := s.getBucket(right, d)

		for leftBucket <= bucket && left < bucketFrom {
			if leftBucket == bucket {
				s.radix.Swap(left, slot)
				slot++
			} else {
				left++
			}
			leftBucket = s.getBucket(left, d)
		}

		for rightBucket >= bucket && right >= bucketTo {
			if rightBucket == bucket {
				s.radix.Swap(right, slot)
				slot++
			} else {
				right--
			}
			rightBucket = s.getBucket(right, d)
		}

		if left < bucketFrom && right >= bucketTo {
			s.radix.Swap(left, right)
			left++
			right--
		} else {
			break
		}
	}
}

// Build an histogram of the k-th characters of values occurring between offsets from and to, using getBucket.
func (s *radixSelector) buildHistogram(from, to, idx int, histogram []int) {
	for i := from; i < to; i++ {
		histogram[s.getBucket(i, idx)]++
	}
}

// Return a number for the k-th character between 0 and HISTOGRAM_SIZE.
func (s *radixSelector) getBucket(i, k int) int {
	return s.radix.ByteAt(i, k) + 1
}

func (s *radixSelector) getFallbackSelector(d int) *introSelector {
	return &introSelector{selector: &fallbackSelector{
		radixSelector: s,
		skipBytes:     d,
	}}
}

var _ IntroSelector = &fallbackSelector{}

type fallbackSelector struct {
	*radixSelector
	skipBytes int
}

func (f *fallbackSelector) Swap(i, j int) {
	f.radix.Swap(i, j)
}

func (f *fallbackSelector) Compare(i, j int) int {
	for idx := f.skipBytes; idx < f.maxLength; idx++ {
		b1 := f.radix.ByteAt(i, idx)
		b2 := f.radix.ByteAt(j, idx)
		if b1 == -1 {
			break
		}

		if b1 != b2 {
			return b1 - b2
		}

	}
	return 0
}
