package fst

func binarySearch[T any](fst *FST[T], arc *Arc[T], targetLabel int) (int, error) {
	in, err := fst.GetBytesReader()
	if err != nil {
		return 0, err
	}

	low := arc.ArcIdx()
	mid := 0
	high := int(arc.NumArcs() - 1)

	for low <= high {
		mid = (low + high) >> 1
		if err := in.SetPosition(arc.PosArcsStart()); err != nil {
			return 0, err
		}
		if err := in.SkipBytes(arc.BytesPerArc()*mid + 1); err != nil {
			return 0, err
		}
		midLabel, err := fst.ReadLabel(in)
		if err != nil {
			return 0, err
		}
		cmp := midLabel - targetLabel
		if cmp < 0 {
			low = mid + 1
		} else if cmp > 0 {
			high = mid - 1
		} else {
			return mid, nil
		}
	}

	return -1 - low, nil
}
