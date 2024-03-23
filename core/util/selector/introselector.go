package selector

import "math"

type IntroSelector interface {
	Swap(i, j int)

	// Compare entries found in slots i and j.
	// The contract for the returned value is the same as Comparator.compare(Object, Object).
	Compare(i, j int) int
}

type introSelector struct {
	selector IntroSelector
}

func NewIntroSelector(intro IntroSelector) Selector {
	return &introSelector{selector: intro}
}

func newIntroSelector(intro IntroSelector) {

}

func (s *introSelector) SelectK(from, to, k int) {
	maxDepth := 2 * math.Log2(float64(to-from))
	s.quickSelect(from, to, k, int(maxDepth))
}

func (s *introSelector) quickSelect(from, to, k, maxDepth int) {
	if to-from == 1 {
		return
	}

	maxDepth--
	if maxDepth < 0 {
		s.slowSelect(from, to, k)
		return
	}

	mid := (from + to) >> 1 // (from + to) / 2

	// heuristic: we use the median of the values at from, to-1 and mid as a pivot
	if s.selector.Compare(from, to-1) > 0 {
		s.selector.Swap(from, to-1)
	}

	if s.selector.Compare(to-1, mid) > 0 {
		s.selector.Swap(to-1, mid)
		if s.selector.Compare(from, to-1) > 0 {
			s.selector.Swap(from, to-1)
		}
	}

	//s.selector.SetPivot(to - 1)

	idx := to - 1

	left, right := from+1, to-2

	for {
		for s.selector.Compare(idx, left) > 0 {
			left++
		}

		for left < right && s.selector.Compare(idx, right) <= 0 {
			right--
		}

		if left < right {
			s.selector.Swap(left, right)
			right--
		} else {
			break
		}
	}
	s.selector.Swap(left, to-1)

	if left == k {
		return
	} else if left < k {
		s.quickSelect(left+1, to, k, maxDepth)
	} else {
		s.quickSelect(from, left, k, maxDepth)
	}
}

func (s *introSelector) slowSelect(from, to, k int) int {
	return s.medianOfMediansSelect(from, to-1, k)
}

func (s *introSelector) medianOfMediansSelect(left, right, k int) int {
	for left != right {
		// Defensive check, this is also checked in the calling
		// method. Including here so this method can be used
		// as a self contained quickSelect implementation.

		pivotIndex := s.pivot(left, right)
		pivotIndex = s.partition(left, right, k, pivotIndex)
		if k == pivotIndex {
			return k
		} else if k < pivotIndex {
			right = pivotIndex - 1
		} else {
			left = pivotIndex + 1
		}
	}
	return left
}

func (s *introSelector) partition(left, right, k, pivotIndex int) int {
	//s.selector.SetPivot(pivotIndex)
	s.selector.Swap(pivotIndex, right)
	storeIndex := left
	for i := left; i < right; i++ {
		if s.selector.Compare(right, i) > 0 {
			s.selector.Swap(storeIndex, i)
			storeIndex++
		}
	}

	storeIndexEq := storeIndex
	for i := storeIndex; i < right; i++ {
		if s.selector.Compare(right, i) == 0 {
			s.selector.Swap(storeIndexEq, i)
			storeIndexEq++
		}
	}

	s.selector.Swap(right, storeIndexEq)
	if k < storeIndex {
		return storeIndex
	}
	if k <= storeIndexEq {
		return k
	}
	return storeIndexEq
}

func (s *introSelector) pivot(left int, right int) int {
	if right-left < 5 {
		pivotIndex := s.partition5(left, right)
		return pivotIndex
	}

	for i := left; i <= right; i += 5 {
		subRight := i + 4
		if subRight > right {
			subRight = right
		}
		median5 := s.partition5(i, subRight)
		s.selector.Swap(median5, left+((i-left)/5))
	}
	mid := ((right - left) / 10) + left + 1
	to := left + ((right - left) / 5)
	return s.medianOfMediansSelect(left, to, mid)
}

// selects the median of a group of at most five elements,
// implemented using insertion sort. Efficient due to
// bounded nature of data set.
func (s *introSelector) partition5(left int, right int) int {
	i := left + 1
	for i <= right {
		j := i
		for j > left && s.selector.Compare(j-1, j) > 0 {
			s.selector.Swap(j-1, j)
			j--
		}
		i++
	}
	return (left + right) >> 1
}
