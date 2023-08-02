package fst

import "context"

func binarySearch(ctx context.Context, fst *FST, arc *Arc, targetLabel int) (int, error) {
	in, err := fst.GetBytesReader()
	if err != nil {
		return 0, err
	}

	low := arc.ArcIdx()
	mid := 0
	high := arc.NumArcs() - 1

	for low <= high {
		mid = (low + high) >> 1
		if err := in.SetPosition(arc.PosArcsStart()); err != nil {
			return 0, err
		}
		if err := in.SkipBytes(ctx, arc.BytesPerArc()*mid+1); err != nil {
			return 0, err
		}
		midLabel, err := fst.ReadLabel(ctx, in)
		if err != nil {
			return 0, err
		}
		cmp := midLabel - targetLabel
		if cmp == 0 {
			return mid, nil
		}

		if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return -1 - low, nil
}
