package fst

// Enum
// Can next() and advance() through the terms in an FST
// lucene.experimental
type Enum[T PairAble] struct {
	fst    *FST[T]
	arcs   []*Arc[T] //
	output []T       // outputs are cumulative

	noOutput     T
	fstReader    BytesReader
	upto         int
	targetLength int

	GetTargetLabel  func() (int, error)
	GetCurrentLabel func() (int, error)
	SetCurrentLabel func(label int) error
	Grow            func() error
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

func NewFstEnum[T PairAble](fst *FST[T]) (*Enum[T], error) {
	reader, err := fst.GetBytesReader()
	if err != nil {
		return nil, err
	}

	noOutput := fst.outputs.GetNoOutput()

	enum := &Enum[T]{
		fst:       fst,
		fstReader: reader,
		noOutput:  noOutput,
		arcs:      make([]*Arc[T], 10),
		output:    make([]T, 10),
	}

	enum.output[0] = noOutput

	for i := range enum.arcs {
		enum.arcs[i] = &Arc[T]{}
	}

	if _, err := fst.GetFirstArc(enum.getArc(0)); err != nil {
		return nil, err
	}

	return enum, nil
}

// Rewinds enum state to match the shared prefix between current term and target term
// 倒回枚举状态，以匹配当前term和目标term之间的共享前缀
func (f *Enum[T]) rewindPrefix() error {
	if f.upto == 0 {
		f.upto = 1
		_, err := f.fst.ReadFirstTargetArc(f.getArc(0), f.getArc(1), f.fstReader)
		return err
	}

	currentLimit := f.upto
	f.upto = 1

	for f.upto < currentLimit && f.upto < f.targetLength+1 {
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
			arc := f.getArc(f.upto)
			if _, err := f.fst.ReadFirstTargetArc(f.arcs[f.upto-1], arc, f.fstReader); err != nil {
				return err
			}
			break
		}

		f.upto++
	}

	return nil
}

func (f *Enum[T]) doNext() error {
	//System.out.println("FE: next upto=" + upto);
	if f.upto == 0 {
		//System.out.println("  init");
		f.upto = 1
		_, err := f.fst.ReadFirstTargetArc(f.getArc(0), f.getArc(1), f.fstReader)
		if err != nil {
			return err
		}
	} else {
		// pop
		// System.out.println("  check pop curArc target=" + arcs[upto].target + " label=" + arcs[upto].label + " isLast?=" + arcs[upto].isLast());
		for f.arcs[f.upto].IsLast() {
			f.upto--
			if f.upto == 0 {
				return nil
			}
		}
		_, err := f.fst.ReadNextArc(f.arcs[f.upto], f.fstReader)
		if err != nil {
			return err
		}
	}

	return f.pushFirst()
}

// Seeks to smallest term that's >= target.
func (f *Enum[T]) doSeekCeil() error {

	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	if err := f.rewindPrefix(); err != nil {
		return err
	}

	arc := f.getArc(f.upto)

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
				// assert arc.nodeFlags() == Fst.ARCS_FOR_BINARY_SEARCH;
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

func (f *Enum[T]) doSeekCeilArrayDirectAddressing(
	arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {

	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex >= int(arc.NumArcs()) {
		// Target is beyond the last arc, out of label range.
		// Dead end (target is after the last arc);
		// rollback to last fork then push
		f.upto--
		for {
			if f.upto == 0 {
				return nil, nil
			}
			prevArc := f.getArc(f.upto)
			//System.out.println("  rollback upto=" + upto + " arc.label=" + prevArc.label + " isLast?=" + prevArc.isLast());
			if !prevArc.IsLast() {
				f.fst.ReadNextArc(prevArc, f.fstReader)
				f.pushFirst()
				return nil, nil
			}
			f.upto--
		}
	} else {
		if targetIndex < 0 {
			targetIndex = -1
		} else if ok, err := (IsBitSet(targetIndex, arc, in)); ok && err == nil {
			f.fst.ReadArcByDirectAddressing(arc, in, targetIndex)
			//assert arc.label() == targetLabel;
			// found -- copy pasta from below
			f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
			if targetLabel == END_LABEL {
				return nil, nil
			}
			f.SetCurrentLabel(arc.Label())
			f.incr()
			return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
		}
		// Not found, return the next arc (ceil).
		ceilIndex, err := NextBitSet(targetIndex, arc, in)
		if err != nil {
			return nil, err
		}
		//assert ceilIndex != -1;
		f.fst.ReadArcByDirectAddressing(arc, in, ceilIndex)
		//assert arc.label() > targetLabel;
		f.pushFirst()
		return nil, nil
	}
}

func (f *Enum[T]) doSeekCeilArrayPacked(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// The array is packed -- use binary search to find the target.
	idx, err := binarySearch(f.fst, arc, targetLabel)
	if err != nil {
		return nil, err
	}

	if (idx) >= 0 {
		// Match
		f.fst.ReadArcByIndex(arc, in, idx)
		//assert arc.arcIdx() == idx;
		//assert arc.label() == targetLabel: "arc.label=" + arc.label() + " vs targetLabel=" + targetLabel + " mid=" + idx;
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if targetLabel == END_LABEL {
			return nil, err
		}
		err := f.SetCurrentLabel(arc.Label())
		if err != nil {
			return nil, err
		}
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	}

	idx = -1 - idx
	if idx == int(arc.NumArcs()) {
		// Dead end
		_, err := f.fst.ReadArcByIndex(arc, in, idx-1)
		if err != nil {
			return nil, err
		}
		//assert arc.isLast();
		// Dead end (target is after the last arc);
		// rollback to last fork then push
		f.upto--
		for {
			if f.upto == 0 {
				return nil, nil
			}
			prevArc := f.getArc(f.upto)
			//System.out.println("  rollback upto=" + upto + " arc.label=" + prevArc.label + " isLast?=" + prevArc.isLast());
			if !prevArc.IsLast() {
				_, err := f.fst.ReadNextArc(prevArc, f.fstReader)
				if err != nil {
					return nil, err
				}
				err = f.pushFirst()
				if err != nil {
					return nil, err
				}
				return nil, nil
			}
			f.upto--
		}
	} else {
		// ceiling - arc with least higher label
		_, err := f.fst.ReadArcByIndex(arc, in, idx)
		if err != nil {
			return nil, err
		}
		//assert arc.label() > targetLabel;
		err = f.pushFirst()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

// Seeks to largest term that's <= target.
func (f *Enum[T]) doSeekFloor() error {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton
	//System.out.println("FE: seek floor upto=" + upto);

	// Save CPU by starting at the end of the shared prefix
	// b/w our current term & the target:
	err := f.rewindPrefix()
	if err != nil {
		return err
	}

	//System.out.println("FE: after rewind upto=" + upto);

	arc := f.getArc(f.upto)

	//System.out.println("FE: init targetLabel=" + targetLabel);

	// Now scan forward, matching the new suffix of the target
	for arc != nil {
		//System.out.println("  cycle upto=" + upto + " arc.label=" + arc.label + " (" + (char) arc.label + ") targetLabel=" + targetLabel + " isLast?=" + arc.isLast() + " bba=" + arc.bytesPerArc);
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
				arc, err = f.doSeekFloorArrayDirectAddressing(arc, targetLabel, in)
				if err != nil {
					return err
				}
			} else {
				//assert arc.nodeFlags() == Fst.ARCS_FOR_BINARY_SEARCH;
				arc, err = f.doSeekFloorArrayPacked(arc, targetLabel, in)
				if err != nil {
					return err
				}
			}
		} else {
			arc, err = f.doSeekFloorList(arc, targetLabel)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Enum[T]) doSeekCeilList(arc *Arc[T], targetLabel int) (*Arc[T], error) {
	var err error
	// Arcs are not array'd -- must do linear scan:
	if arc.Label() == targetLabel {
		// recurse
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if err != nil {
			return nil, err
		}
		if targetLabel == END_LABEL {
			return nil, nil
		}
		err := f.SetCurrentLabel(arc.Label())
		if err != nil {
			return nil, err
		}
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	} else if arc.Label() > targetLabel {
		err := f.pushFirst()
		if err != nil {
			return nil, err
		}
		return nil, nil
	} else if arc.IsLast() {
		// Dead end (target is after the last arc);
		// rollback to last fork then push
		f.upto--
		for {
			if f.upto == 0 {
				return nil, nil
			}
			prevArc := f.getArc(f.upto)
			//System.out.println("  rollback upto=" + upto + " arc.label=" + prevArc.label + " isLast?=" + prevArc.isLast());
			if !prevArc.IsLast() {
				_, err := f.fst.ReadNextArc(prevArc, f.fstReader)
				if err != nil {
					return nil, err
				}
				err = f.pushFirst()
				if err != nil {
					return nil, err
				}
				return nil, nil
			}
			f.upto--
		}
	} else {
		// keep scanning
		//System.out.println("    next scan");
		_, err := f.fst.ReadNextArc(arc, f.fstReader)
		if err != nil {
			return nil, err
		}
	}
	return arc, nil
}

func (f *Enum[T]) doSeekFloorArrayDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex < 0 {
		// Before first arc.
		return f.backtrackToFloorArc(arc, targetLabel, in)
	} else if targetIndex >= int(arc.NumArcs()) {
		// After last arc.
		f.fst.ReadLastArcByDirectAddressing(arc, in)
		//assert arc.label() < targetLabel;
		//assert arc.isLast();
		f.pushLast()
		return nil, nil
	} else {
		var err error
		// Within label range.
		if ok, _ := IsBitSet(targetIndex, arc, in); ok {
			f.fst.ReadArcByDirectAddressing(arc, in, targetIndex)
			//assert arc.label() == targetLabel;
			// found -- copy pasta from below
			f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
			if err != nil {
				return nil, err
			}
			if targetLabel == END_LABEL {
				return nil, nil
			}
			f.SetCurrentLabel(arc.Label())
			f.incr()
			return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
		}
		// Scan backwards to find a floor arc.
		floorIndex, err := PreviousBitSet(targetIndex, arc, in)
		if err != nil {
			return nil, err
		}
		//assert floorIndex != -1;
		f.fst.ReadArcByDirectAddressing(arc, in, floorIndex)
		//assert arc.label() < targetLabel;
		//assert arc.isLast() || fst.readNextArcLabel(arc, in) > targetLabel;
		f.pushLast()
		return nil, nil
	}
}

// Backtracks until it finds a node which first arc is before our target label.` Then on the node, finds the arc just before the targetLabel.
// 返回值:
// null to continue the seek floor recursion loop.
func (f *Enum[T]) backtrackToFloorArc(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	var err error
	for {
		// First, walk backwards until we find a node which first arc is before our target label.
		f.fst.ReadFirstTargetArc(f.getArc(f.upto-1), arc, f.fstReader)
		if arc.Label() < targetLabel {
			// Then on this node, find the arc just before the targetLabel.
			if !arc.IsLast() {
				if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
					if arc.NodeFlags() == ARCS_FOR_BINARY_SEARCH {
						f.findNextFloorArcBinarySearch(arc, targetLabel, in)
					} else {
						//assert arc.nodeFlags() == Fst.ARCS_FOR_DIRECT_ADDRESSING;
						f.findNextFloorArcDirectAddressing(arc, targetLabel, in)
					}
				} else {

					for {
						if !arc.IsLast() {
							if n, _ := f.fst.readNextArcLabel(arc, in); n < targetLabel {
								f.fst.ReadNextArc(arc, f.fstReader)
								continue
							}
						}
						break
					}
				}
			}
			//assert arc.label() < targetLabel;
			//assert arc.isLast() || fst.readNextArcLabel(arc, in) >= targetLabel;
			f.pushLast()
			return nil, nil
		}
		f.upto--
		if f.upto == 0 {
			return nil, nil
		}
		targetLabel, err = f.GetTargetLabel()
		if err != nil {
			return nil, err
		}
		arc = f.getArc(f.upto)
	}
}

// Finds and reads an arc on the current node which label is strictly less than the given label.
// Skips the first arc, finds next floor arc; or none if the floor arc is the first arc itself
// (in this case it has already been read).
// Precondition: the given arc is the first arc of the node.
func (f *Enum[T]) findNextFloorArcDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) error {
	//assert arc.nodeFlags() == Fst.ARCS_FOR_DIRECT_ADDRESSING;
	//assert arc.label() != Fst.END_LABEL;
	//assert arc.label() == arc.firstLabel();
	if arc.NumArcs() > 1 {
		targetIndex := targetLabel - arc.FirstLabel()
		//assert targetIndex >= 0;
		if targetIndex >= int(arc.NumArcs()) {
			// Beyond last arc. Take last arc.
			f.fst.ReadLastArcByDirectAddressing(arc, in)
		} else {
			// Take the preceding arc, even if the target is present.
			floorIndex, err := PreviousBitSet(targetIndex, arc, in)
			if err != nil {
				return err
			}
			if floorIndex > 0 {
				f.fst.ReadArcByDirectAddressing(arc, in, floorIndex)
			}
		}
	}
	return nil
}

// Same as findNextFloorArcDirectAddressing for binary search node.
func (f *Enum[T]) findNextFloorArcBinarySearch(arc *Arc[T], targetLabel int, in BytesReader) error {
	//assert arc.nodeFlags() == Fst.ARCS_FOR_BINARY_SEARCH;
	//assert arc.label() != Fst.END_LABEL;
	//assert arc.arcIdx() == 0;
	if arc.NumArcs() > 1 {
		idx, err := binarySearch(f.fst, arc, targetLabel)
		if err != nil {
			return err
		}
		//assert idx != -1;
		if idx > 1 {
			f.fst.ReadArcByIndex(arc, in, idx-1)
		} else if idx < -2 {
			f.fst.ReadArcByIndex(arc, in, -2-idx)
		}
	}
	return nil
}

func (f *Enum[T]) doSeekFloorArrayPacked(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// Arcs are fixed array -- use binary search to find the target.
	idx, err := binarySearch(f.fst, arc, targetLabel)
	if err != nil {
		return nil, err
	}

	if idx >= 0 {
		// Match -- recurse
		//System.out.println("  match!  arcIdx=" + idx);
		f.fst.ReadArcByIndex(arc, in, idx)
		//assert arc.arcIdx() == idx;
		//assert arc.label() == targetLabel: "arc.label=" + arc.label() + " vs targetLabel=" + targetLabel + " mid=" + idx;
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if err != nil {
			return nil, err
		}
		if targetLabel == END_LABEL {
			return nil, nil
		}
		f.SetCurrentLabel(arc.Label())
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	} else if idx == -1 {
		// Before first arc.
		return f.backtrackToFloorArc(arc, targetLabel, in)
	} else {
		// There is a floor arc; idx will be (-1 - (floor + 1)).
		f.fst.ReadArcByIndex(arc, in, -2-idx)
		//assert arc.isLast() || fst.readNextArcLabel(arc, in) > targetLabel;
		//assert arc.label() < targetLabel: "arc.label=" + arc.label() + " vs targetLabel=" + targetLabel;
		f.pushLast()
		return nil, nil
	}
}

func (f *Enum[T]) doSeekFloorList(arc *Arc[T], targetLabel int) (*Arc[T], error) {
	var err error
	if arc.Label() == targetLabel {
		// Match -- recurse
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if err != nil {
			return nil, err
		}
		if targetLabel == END_LABEL {
			return nil, nil
		}
		f.SetCurrentLabel(arc.Label())
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	} else if arc.Label() > targetLabel {
		// TODO: if each arc could somehow read the arc just
		// before, we can save this re-scan.  The ceil case
		// doesn't need this because it reads the next arc
		// instead:
		for {
			// First, walk backwards until we find a first arc
			// that's before our target label:
			f.fst.ReadFirstTargetArc(f.getArc(f.upto-1), arc, f.fstReader)
			if arc.Label() < targetLabel {
				// Then, scan forwards to the arc just before
				// the targetLabel:
				for {
					if !arc.IsLast() {
						if n, _ := f.fst.readNextArcLabel(arc, f.fstReader); n < targetLabel {
							f.fst.ReadNextArc(arc, f.fstReader)
							continue
						}
					}
					break
				}

				f.pushLast()
				return nil, nil
			}
			f.upto--
			if f.upto == 0 {
				return nil, nil
			}
			targetLabel, err = f.GetTargetLabel()
			if err != nil {
				return nil, err
			}
			arc = f.getArc(f.upto)
		}
	} else if !arc.IsLast() {
		//System.out.println("  check next label=" + fst.readNextArcLabel(arc) + " (" + (char) fst.readNextArcLabel(arc) + ")");
		if n, _ := f.fst.readNextArcLabel(arc, f.fstReader); n > targetLabel {
			f.pushLast()
			return nil, nil
		} else {
			// keep scanning
			return f.fst.ReadNextArc(arc, f.fstReader)
		}
	} else {
		f.pushLast()
		return nil, nil
	}
}

// DoSeekExact Seeks to exactly target term.
func (f *Enum[T]) DoSeekExact() (bool, error) {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	if err := f.rewindPrefix(); err != nil {
		return false, err
	}

	arc := f.getArc(f.upto - 1)
	targetLabel, err := f.GetTargetLabel()
	if err != nil {
		return false, err
	}

	fstReader, err := f.fst.GetBytesReader()
	if err != nil {
		return false, err
	}

	for {
		nextArc, err := f.fst.FindTargetArc(targetLabel, arc, f.getArc(f.upto), fstReader)
		if err != nil {
			return false, err
		}

		if nextArc == nil {
			_, err := f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), fstReader)
			if err != nil {
				return false, err
			}
			return false, nil
		}

		// Match -- recurse:
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], nextArc.Output())
		if err != nil {
			return false, err
		}

		if targetLabel == END_LABEL {
			return true, nil
		}

		if err = f.SetCurrentLabel(targetLabel); err != nil {
			return false, err
		}
		f.incr()
		targetLabel, err = f.GetTargetLabel()
		if err != nil {
			return false, err
		}
		arc = nextArc
	}
}

// Appends current arc, and then recurses from its target,
// appending first arc all the way to the final node
func (f *Enum[T]) pushFirst() error {

	arc := f.arcs[f.upto]

	var err error
	for {
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
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

		nextArc := f.getArc(f.upto)
		if _, err := f.fst.ReadFirstTargetArc(arc, nextArc, f.fstReader); err != nil {
			return err
		}
		arc = nextArc
	}
	return nil
}

// Recurses from current arc, appending last arc all the
// way to the first final node
func (f *Enum[T]) pushLast() error {
	arc := f.arcs[f.upto]
	//assert arc != null;

	var err error
	for {
		f.SetCurrentLabel(arc.Label())
		f.output[f.upto], err = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if err != nil {
			return err
		}
		if arc.Label() == END_LABEL {
			// Final node
			break
		}
		f.incr()

		arc, err = f.fst.readLastTargetArc(arc, f.getArc(f.upto), f.fstReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Enum[T]) getArc(idx int) *Arc[T] {
	if len(f.arcs) <= idx {
		f.arcs = append(f.arcs, &Arc[T]{})
	}
	return f.arcs[idx]
}

func (f *Enum[T]) incr() {
	f.upto++
	f.Grow()
	if len(f.arcs) <= f.upto {
		f.arcs = append(f.arcs, &Arc[T]{})
	}
	if len(f.output) <= f.upto {
		f.output = append(f.output, f.noOutput)
	}
}
