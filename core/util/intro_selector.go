package util

// IntroSelector Implementation of the quick select algorithm.
// It uses the median of the first, middle and last values as a buffer and falls back to a median of
// medians when the number of recursion levels exceeds 2 lg(n), as a consequence it runs in linear
// time on average.
// lucene.internal
//type IntroSelector interface {
//	Selector
//
//	// SetPivot Save the value at slot i so that it can later be used as a buffer, see fnComparePivot(int).
//	SetPivot(i int)
//
//	// ComparePivot Compare the buffer with the slot at j, similarly to Compare(i, j).
//	ComparePivot(i int) int
//}

type IntroSelectorConfig struct {
	FnSwap         func(i, j int)
	FnSetPivot     func(i int)
	FnComparePivot func(i int) int
	FnCompare      func(i, j int) int
}

var _ Selector = &IntroSelector{}

type IntroSelector struct {
	fnSwap         func(i, j int)
	fnSetPivot     func(i int)
	fnComparePivot func(i int) int
	fnCompare      func(i, j int) int
}

func (r *IntroSelector) Swap(i, j int) {
	r.fnSwap(i, j)
}

func NewIntroSelector(cfg *IntroSelectorConfig) *IntroSelector {
	return &IntroSelector{
		fnSwap:         cfg.FnSwap,
		fnSetPivot:     cfg.FnSetPivot,
		fnComparePivot: cfg.FnComparePivot,
		fnCompare:      cfg.FnCompare,
	}
}

func (r *IntroSelector) Select(from, to, k int) {
	SelectorCheckArgs(from, to, k)
	maxDepth := 2 * log(to-from, 2)
	r.quickSelect(from, to, k, maxDepth)
}

func (r *IntroSelector) slowSelect(from, to, k int) int {
	return r.medianOfMediansSelect(from, to-1, k)
}

func (r *IntroSelector) medianOfMediansSelect(left, right, k int) int {
	for {
		// Defensive check, this is also checked in the calling
		// method. Including here so this method can be used
		// as a self contained quickSelect implementation.
		if left == right {
			return left
		}
		pivotIndex := r.pivot(left, right)
		pivotIndex = r.partition(left, right, k, pivotIndex)
		if k == pivotIndex {
			return k
		} else if k < pivotIndex {
			right = pivotIndex - 1
		} else {
			left = pivotIndex + 1
		}

		if left != right {
			continue
		} else {
			break
		}
	}
	return left
}

func (r *IntroSelector) partition(left, right, k, pivotIndex int) int {
	r.fnSetPivot(pivotIndex)
	r.fnSwap(pivotIndex, right)
	storeIndex := left
	for i := left; i < right; i++ {
		if r.fnComparePivot(i) > 0 {
			r.fnSwap(storeIndex, i)
			storeIndex++
		}
	}
	storeIndexEq := storeIndex
	for i := storeIndex; i < right; i++ {
		if r.fnComparePivot(i) == 0 {
			r.fnSwap(storeIndexEq, i)
			storeIndexEq++
		}
	}
	r.fnSwap(right, storeIndexEq)
	if k < storeIndex {
		return storeIndex
	} else if k <= storeIndexEq {
		return k
	}
	return storeIndexEq
}

func (r *IntroSelector) pivot(left, right int) int {
	if right-left < 5 {
		pivotIndex := r.partition5(left, right)
		return pivotIndex
	}

	for i := left; i <= right; i = i + 5 {
		subRight := i + 4
		if subRight > right {
			subRight = right
		}
		median5 := r.partition5(i, subRight)
		r.fnSwap(median5, left+((i-left)/5))
	}
	mid := ((right - left) / 10) + left + 1
	to := left + ((right - left) / 5)
	return r.medianOfMediansSelect(left, to, mid)
}

// selects the median of a group of at most five elements,
// implemented using insertion sort. Efficient due to
// bounded nature of data set.
func (r *IntroSelector) partition5(left, right int) int {
	i := left + 1
	for i <= right {
		j := i
		for j > left && r.compare(j-1, j) > 0 {
			r.fnSwap(j-1, j)
			j--
		}
		i++
	}
	return (left + right) >> 1
}

func (r *IntroSelector) quickSelect(from, to, k, maxDepth int) {
	//assert from <= k;
	//assert k < to;
	if to-from == 1 {
		return
	}
	maxDepth--
	if maxDepth < 0 {
		r.slowSelect(from, to, k)
		return
	}

	mid := (from + to) >> 1
	// heuristic: we use the median of the values at from, to-1 and mid as a buffer
	if r.compare(from, to-1) > 0 {
		r.fnSwap(from, to-1)
	}
	if r.compare(to-1, mid) > 0 {
		r.fnSwap(to-1, mid)
		if r.compare(from, to-1) > 0 {
			r.fnSwap(from, to-1)
		}
	}

	r.fnSetPivot(to - 1)

	left := from + 1
	right := to - 2

	for {
		for r.fnComparePivot(left) > 0 {
			left++
		}

		for left < right && r.fnComparePivot(right) <= 0 {
			right--
		}

		if left < right {
			r.fnSwap(left, right)
			right--
		} else {
			break
		}
	}
	r.fnSwap(left, to-1)

	if left == k {
		return
	} else if left < k {
		r.quickSelect(left+1, to, k, maxDepth)
	} else {
		r.quickSelect(from, left, k, maxDepth)
	}
}

func (r *IntroSelector) compare(i, j int) int {
	if r.fnCompare != nil {
		return r.fnCompare(i, j)
	}
	r.fnSetPivot(i)
	return r.fnComparePivot(j)
}

func log(x int, base int) int {
	ret := 0
	for x >= base {
		x /= base
		ret++
	}
	return ret
}
