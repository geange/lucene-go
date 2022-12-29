package fst

import "errors"

//	doFloor controls the behavior of advance: if it's true doFloor is true,
//
// advance positions to the biggest term before target.
type FstEnum[T PairAble] struct {
	fst    *FST[T]
	arcs   []*Arc[T]
	output []T

	noOutput  T
	fstReader BytesReader
	//upto         int
	targetLength int

	GetTargetLabel  func() (int, error)
	GetCurrentLabel func() (int, error)
	SetCurrentLabel func(label int) error
}

func NewArcOutputs[T PairAble]() *ArcOutputs[T] {
	return &ArcOutputs[T]{
		Arc: &Arc[T]{},
	}
}

type ArcOutputs[T PairAble] struct {
	Arc    *Arc[T]
	Output T
}

func NewFstEnum[T PairAble](fst *FST[T]) (*FstEnum[T], error) {
	reader, err := fst.GetBytesReader()
	if err != nil {
		return nil, err
	}

	noOutput := fst.outputs.GetNoOutput()

	enum := &FstEnum[T]{
		fst:       fst,
		fstReader: reader,
		noOutput:  noOutput,
		arcs:      []*Arc[T]{&Arc[T]{}},
		output:    []T{noOutput},
	}

	if _, err := fst.GetFirstArc(enum.arcs[0]); err != nil {
		return nil, err
	}

	return enum, nil
}

// Rewinds enum state to match the shared prefix between current term and target term
// 倒回枚举状态，以匹配当前term和目标term之间的共享前缀
func (f *FstEnum[T]) rewindPrefix() error {
	if len(f.arcs) == 1 {
		f.arcs = append(f.arcs, &Arc[T]{})
		_, err := f.fst.ReadFirstTargetArc(f.arcs[0], f.arcs[1], f.fstReader)
		return err
	}

	currentLimit := len(f.arcs) - 1
	i := 1

	for ; i < currentLimit && i < f.targetLength+1; i++ {
		label1, err := f.GetCurrentLabel()
		if err != nil {
			return err
		}
		label2, err := f.GetTargetLabel()
		if err != nil {
			return err
		}

		cmp := label1 - label2
		if cmp < 0 {
			// seek forward
			// 向前搜素
			break
		}

		if cmp > 0 {
			// seek backwards -- reset this arc to the first arc
			// 向后搜索
			if len(f.arcs) <= i {
				f.arcs = append(f.arcs, &Arc[T]{})
			}
			if _, err := f.fst.ReadFirstTargetArc(f.arcs[i-1], f.arcs[i], f.fstReader); err != nil {
				return err
			}
			break
		}
	}

	if i <= currentLimit {
		f.arcs = f.arcs[:i+1]
	}

	return nil
}

func (f *FstEnum[T]) doNext() error {
	//System.out.println("FE: next upto=" + upto);
	if len(f.arcs) == 1 {
		//System.out.println("  init");
		f.arcs = append(f.arcs, &Arc[T]{})
		f.fst.ReadFirstTargetArc(f.getArc(0), f.getArc(1), f.fstReader)
	} else {
		// pop
		//System.out.println("  check pop curArc target=" + arcs[upto].target + " label=" + arcs[upto].label + " isLast?=" + arcs[upto].isLast());
		i := 0
		for i = len(f.arcs); i >= 0; i-- {
			if !f.arcs[i].IsFinal() {
				break
			}
		}
		f.arcs = f.arcs[:i+1]
		f.fst.ReadNextArc(f.lastArc(), f.fstReader)
	}

	return f.pushFirst()
}

// Seeks to smallest term that's >= target.
func (f *FstEnum[T]) doSeekCeil() error {

	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	if err := f.rewindPrefix(); err != nil {
		return err
	}

	arc := f.lastArc()

	for arc != nil {
		targetLabel, err := f.GetTargetLabel()
		if err != nil {
			return err
		}

		if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
			// Arcs are in an array
			in, err := f.fst.GetBytesReader()
			if err != nil {
				return err
			}
			if arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING {
				arc, err = f.doSeekCeilArrayDirectAddressing(arc, targetLabel, in)
				if err != nil {
					return err
				}
			} else {
				// assert arc.nodeFlags() == FST.ARCS_FOR_BINARY_SEARCH;
				arc, err = f.doSeekCeilArrayPacked(arc, targetLabel, in)
				if err != nil {
					return err
				}
			}
		} else {
			arc, err = f.doSeekCeilList(arc, targetLabel)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FstEnum[T]) doSeekCeilArrayDirectAddressing(
	arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {

	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	// targetIndex := targetLabel - arc.FirstLabel()

	panic("")

}

func (f *FstEnum[T]) doSeekCeilArrayPacked(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {

	panic("")
}

// Seeks to largest term that's <= target.
func (f *FstEnum[T]) doSeekFloor() error {
	panic("")
}

func (f *FstEnum[T]) doSeekCeilList(arc *Arc[T], targetLabel int) (*Arc[T], error) {

	panic("")
}

func (f *FstEnum[T]) doSeekFloorArrayDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	panic("")
}

// Backtracks until it finds a node which first arc is before our target label.` Then on the node, finds the arc just before the targetLabel.
// 返回值:
// null to continue the seek floor recursion loop.
func (f *FstEnum[T]) backtrackToFloorArc(arc *Arc[T], targetLabel int, in BytesReader) (*FST[T], error) {
	panic("")
}

// Finds and reads an arc on the current node which label is strictly less than the given label.
// Skips the first arc, finds next floor arc; or none if the floor arc is the first arc itself
// (in this case it has already been read).
// Precondition: the given arc is the first arc of the node.
func (f *FstEnum[T]) findNextFloorArcDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) error {
	panic("")
}

// Same as findNextFloorArcDirectAddressing for binary search node.
func (f *FstEnum[T]) findNextFloorArcBinarySearch(arc *Arc[T], targetLabel int, in BytesReader) error {
	panic("")
}

func (f *FstEnum[T]) doSeekFloorArrayPacked(arc *Arc[T], targetLabel int, in BytesReader) (*FST[T], error) {
	panic("")
}

func (f *FstEnum[T]) doSeekFloorList(arc *Arc[T], targetLabel int) (*Arc[T], error) {
	panic("")
}

// DoSeekExact Seeks to exactly target term.
func (f *FstEnum[T]) DoSeekExact() (bool, error) {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	//if err := f.rewindPrefix(); err != nil {
	//	return false, err
	//}

	arc := f.arcs[0]
	targetLabel, err := f.GetTargetLabel()
	if err != nil {
		return false, err
	}

	fstReader, err := f.fst.GetBytesReader()
	if err != nil {
		return false, err
	}

	for {
		next, err := f.fst.FindTarget(targetLabel, arc, fstReader)
		if err != nil {
			return false, err
		}
		f.arcs = append(f.arcs, next)

		if next == nil {
			next, err := f.fst.ReadFirstTarget(arc, fstReader)
			if err != nil {
				return false, err
			}
			f.arcs = append(f.arcs, next)
			return false, nil
		}

		currentOutput := f.noOutput
		if len(f.output) > 0 {
			currentOutput = f.output[len(f.output)-1]
		}

		newOutput, err := f.fst.outputs.Add(currentOutput, next.Output())
		if err != nil {
			return false, err
		}
		f.output = append(f.output, newOutput)

		if targetLabel == END_LABEL {
			//System.out.println("  return found; upto=" + upto + " output=" + output[upto] + " next=" + next.isLast());
			return true, nil
		}
		if err = f.SetCurrentLabel(targetLabel); err != nil {
			return false, err
		}
		targetLabel, err = f.GetTargetLabel()
		if err != nil {
			return false, err
		}
		arc = next
	}
}

// Appends current arc, and then recurses from its target,
// appending first arc all the way to the final node
func (f *FstEnum[T]) pushFirst() error {
	if len(f.arcs) == 0 {
		return errors.New("arcs size is zero")
	}

	upto := len(f.arcs) - 1

	arc := f.arcs[upto]

	var err error
	for {
		f.output[upto], err = f.fst.outputs.Add(f.output[upto-1], arc.Output())
		if err != nil {
			return err
		}
		if arc.Label() == END_LABEL {
			// Final node
			break
		}

		if err := f.SetCurrentLabel(arc.Label()); err != nil {
			return err
		}

		nextArc := f.getArc(upto)
		if _, err := f.fst.ReadFirstTargetArc(arc, nextArc, f.fstReader); err != nil {
			return err
		}
		arc = nextArc
	}
	return nil
}

// Recurses from current arc, appending last arc all the
// way to the first final node
func (f *FstEnum[T]) pushLast() error {
	panic("")
}

func (f *FstEnum[T]) getArc(idx int) *Arc[T] {
	if len(f.arcs) <= idx {
		f.arcs = append(f.arcs, &Arc[T]{})
	}
	return f.arcs[idx]
}

func (f *FstEnum[T]) lastArc() *Arc[T] {
	return f.arcs[len(f.arcs)-1]
}
