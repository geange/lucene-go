package fst

// Perform a binary search of Arcs encoded as a packed array
// Params: fst – the FST from which to read arc – the starting arc; sibling arcs greater than this will be searched. Usually the first arc in the array. targetLabel – the label to search for
// Returns: the index of the Arc having the target label, or if no Arc has the matching label, -1 - idx), where idx is the index of the Arc with the next highest label, or the total number of arcs if the target label exceeds the maximum.
// Throws: IOException – when the FST reader does
func binarySearch[T any](fst *FST[T], arc *Arc[T], targetLabel int) int {
	in := fst.GetBytesReader()
	low := arc.ArcIdx()
	mid := 0
	high := arc.NumArcs() - 1
	for low <= high {
		mid = (low + high) >> 1
		in.SetPosition(arc.PosArcsStart())
		in.SkipBytes(arc.BytesPerArc()*mid + 1)
		midLabel, err := fst.ReadLabel(in)
		if err != nil {
			return -1
		}
		cmp := midLabel - targetLabel
		if cmp < 0 {
			low = mid + 1
		} else if cmp > 0 {
			high = mid - 1
		} else {
			return mid
		}
	}
	return -1 - low
}
