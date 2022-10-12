package fst

// FSTEnum Can next() and advance() through the terms in an FST
// lucene.experimental
type FSTEnum interface {
	GetTargetLabel() int
	GetCurrentLabel() int
	SetCurrentLabel(label int)
	Grow()
}

type FSTEnumImp[T any] struct {
	Imp FSTEnum

	fst          *FST[T]
	arcs         []*Arc[T]
	output       []*Box[T]
	noOutput     *Box[T]
	fstReader    BytesReader
	upto         int
	targetLength int
}

// Rewinds enum state to match the shared prefix between current term and target term
func (f *FSTEnumImp[T]) rewindPrefix() error {
	if f.upto == 0 {
		//System.out.println("  init");
		f.upto = 1
		_, err := f.fst.ReadFirstTargetArc(f.getArc(0), f.getArc(1), f.fstReader)
		return err
	}
	//System.out.println("  rewind upto=" + upto + " vs targetLength=" + targetLength);

	currentLimit := f.upto
	f.upto = 1
	for f.upto < currentLimit && f.upto <= f.targetLength+1 {
		cmp := f.Imp.GetCurrentLabel() - f.Imp.GetTargetLabel()
		if cmp < 0 {
			// seek forward
			break
		} else if cmp > 0 {
			// seek backwards -- reset this arc to the first arc
			arc := f.getArc(f.upto)
			_, err := f.fst.ReadFirstTargetArc(f.getArc(f.upto-1), arc, f.fstReader)
			if err != nil {
				return err
			}
			//System.out.println("    seek first arc");
			break
		}
		f.upto++
	}
	return nil
}

func (f *FSTEnumImp[T]) doNext() error {
	if f.upto == 0 {
		//System.out.println("  init");
		f.upto = 1
		f.fst.ReadFirstTargetArc(f.getArc(0), f.getArc(1), f.fstReader)
	} else {
		// pop
		for f.arcs[f.upto].IsLast() {
			f.upto--
			if f.upto == 0 {
				//System.out.println("  eof");
				return nil
			}
		}
		f.fst.ReadNextArc(f.arcs[f.upto], f.fstReader)
	}

	return f.pushFirst()
}

// TODO: should we return a status here (SEEK_FOUND / SEEK_NOT_FOUND /
// SEEK_END)?  saves the eq check above?

// Seeks to smallest term that's >= target.
func (f *FSTEnumImp[T]) doSeekCeil() error {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	//System.out.println("FE.seekCeil upto=" + upto);

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	err := f.rewindPrefix()
	if err != nil {
		return err
	}
	//System.out.println("  after rewind upto=" + upto);

	arc := f.getArc(f.upto)
	//System.out.println("  init targetLabel=" + targetLabel);

	// Now scan forward, matching the new suffix of the target
	for arc != nil {
		targetLabel := f.Imp.GetTargetLabel()
		//System.out.println("  cycle upto=" + upto + " arc.label=" + arc.label + " (" + (char) arc.label + ") vs targetLabel=" + targetLabel);
		if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
			// Arcs are in an array
			in := f.fst.GetBytesReader()
			if arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING {
				arc, err = f.doSeekCeilArrayDirectAddressing(arc, targetLabel, in)
			} else {
				arc, err = f.doSeekCeilArrayPacked(arc, targetLabel, in)
			}
		} else {
			arc, err = f.doSeekCeilList(arc, targetLabel)
		}
	}
	return nil
}

func (f *FSTEnumImp[T]) doSeekCeilArrayDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex >= arc.NumArcs() {
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
		if targetIndex < 0 {
			targetIndex = -1
		} else if IsBitSet(targetIndex, arc, in) {
			_, err := f.fst.ReadArcByDirectAddressing(arc, in, targetIndex)
			if err != nil {
				return nil, err
			}

			// found -- copy pasta from below
			f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
			if targetLabel == END_LABEL {
				return nil, nil
			}
			f.Imp.SetCurrentLabel(arc.Label())
			f.incr()
			return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
		}
		// Not found, return the next arc (ceil).
		ceilIndex := NextBitSet(targetIndex, arc, in)

		_, err := f.fst.ReadArcByDirectAddressing(arc, in, ceilIndex)
		if err != nil {
			return nil, err
		}

		err = f.pushFirst()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (f *FSTEnumImp[T]) doSeekCeilArrayPacked(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// The array is packed -- use binary search to find the target.
	idx := binarySearch(f.fst, arc, targetLabel)
	if idx >= 0 {
		// Match
		f.fst.ReadArcByIndex(arc, in, idx)

		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if targetLabel == END_LABEL {
			return nil, nil
		}
		f.Imp.SetCurrentLabel(arc.Label())
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	}
	idx = -1 - idx
	if idx == arc.NumArcs() {
		// Dead end
		f.fst.ReadArcByIndex(arc, in, idx-1)

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
		// Ceiling - arc with least higher label
		f.fst.ReadArcByIndex(arc, in, idx)
		f.pushFirst()
		return nil, nil
	}
}

func (f *FSTEnumImp[T]) doSeekCeilList(arc *Arc[T], targetLabel int) (*Arc[T], error) {
	// Arcs are not array'd -- must do linear scan:
	if arc.Label() == targetLabel {
		// recurse
		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if targetLabel == END_LABEL {
			return nil, nil
		}
		f.Imp.SetCurrentLabel(arc.Label())
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	} else if arc.Label() > targetLabel {
		f.pushFirst()
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
				f.fst.ReadNextArc(prevArc, f.fstReader)
				f.pushFirst()
				return nil, nil
			}
			f.upto--
		}
	} else {
		// keep scanning
		//System.out.println("    next scan");
		f.fst.ReadNextArc(arc, f.fstReader)
	}
	return arc, nil
}

// Todo: should we return a status here (SEEK_FOUND / SEEK_NOT_FOUND /
// SEEK_END)?  saves the eq check above?

// Seeks to largest term that's <= target.
func (f *FSTEnumImp[T]) doSeekFloor() error {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save CPU by starting at the end of the shared prefix
	// b/w our current term & the target:
	err := f.rewindPrefix()
	if err != nil {
		return err
	}

	arc := f.getArc(f.upto)

	// Now scan forward, matching the new suffix of the target
	for arc != nil {
		targetLabel := f.Imp.GetTargetLabel()

		if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
			// Arcs are in an array
			in := f.fst.GetBytesReader()
			if arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING {
				arc, err = f.doSeekFloorArrayDirectAddressing(arc, targetLabel, in)
				if err != nil {
					return err
				}
			} else {
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

func (f *FSTEnumImp[T]) doSeekFloorArrayDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex < 0 {
		// Before first arc.
		return f.backtrackToFloorArc(arc, targetLabel, in)
	} else if targetIndex >= arc.NumArcs() {
		// After last arc.
		f.fst.ReadLastArcByDirectAddressing(arc, in)

		f.pushLast()
		return nil, nil
	} else {
		// Within label range.
		if IsBitSet(targetIndex, arc, in) {
			f.fst.ReadArcByDirectAddressing(arc, in, targetIndex)
			// found -- copy pasta from below
			f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
			if targetLabel == END_LABEL {
				return nil, nil
			}
			f.Imp.SetCurrentLabel(arc.Label())
			f.incr()
			return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
		}
		// Scan backwards to find a floor arc.
		floorIndex := PreviousBitSet(targetIndex, arc, in)
		f.fst.ReadArcByDirectAddressing(arc, in, floorIndex)

		f.pushLast()
		return nil, nil
	}
}

// Backtracks until it finds a node which first arc is before our target label.` Then on the node,
// finds the arc just before the targetLabel.
// Returns: null to continue the seek floor recursion loop.
func (f *FSTEnumImp[T]) backtrackToFloorArc(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
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
						f.findNextFloorArcDirectAddressing(arc, targetLabel, in)
					}
				} else {
					label, err := f.fst.readNextArcLabel(arc, in)
					if err != nil {
						return nil, err
					}
					for !arc.IsLast() && label < targetLabel {
						f.fst.ReadNextArc(arc, f.fstReader)
						label, err = f.fst.readNextArcLabel(arc, in)
						if err != nil {
							return nil, err
						}
					}
				}
			}

			f.pushLast()
			return nil, nil
		}
		f.upto--
		if f.upto == 0 {
			return nil, nil
		}
		targetLabel = f.Imp.GetTargetLabel()
		arc = f.getArc(f.upto)
	}
}

// Finds and reads an arc on the current node which label is strictly less than the given label.
// Skips the first arc, finds next floor arc; or none if the floor arc is the first arc itself
// (in this case it has already been read).
// Precondition: the given arc is the first arc of the node.
func (f *FSTEnumImp[T]) findNextFloorArcDirectAddressing(arc *Arc[T], targetLabel int, in BytesReader) error {
	if arc.NumArcs() > 1 {
		targetIndex := targetLabel - arc.FirstLabel()

		if targetIndex >= arc.NumArcs() {
			// Beyond last arc. Take last arc.
			f.fst.ReadLastArcByDirectAddressing(arc, in)
		} else {
			// Take the preceding arc, even if the target is present.
			floorIndex := PreviousBitSet(targetIndex, arc, in)
			if floorIndex > 0 {
				f.fst.ReadArcByDirectAddressing(arc, in, floorIndex)
			}
		}
	}
	return nil
}

// Same as findNextFloorArcDirectAddressing for binary search node.
func (f *FSTEnumImp[T]) findNextFloorArcBinarySearch(arc *Arc[T], targetLabel int, in BytesReader) error {
	if arc.NumArcs() > 1 {
		idx := binarySearch(f.fst, arc, targetLabel)

		if idx > 1 {
			_, err := f.fst.ReadArcByIndex(arc, in, idx-1)
			if err != nil {
				return err
			}
		} else if idx < -2 {
			_, err := f.fst.ReadArcByIndex(arc, in, -2-idx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FSTEnumImp[T]) doSeekFloorArrayPacked(arc *Arc[T], targetLabel int, in BytesReader) (*Arc[T], error) {
	// Arcs are fixed array -- use binary search to find the target.
	idx := binarySearch(f.fst, arc, targetLabel)

	if idx >= 0 {
		// Match -- recurse
		//System.out.println("  match!  arcIdx=" + idx);
		f.fst.ReadArcByIndex(arc, in, idx)

		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if targetLabel == END_LABEL {
			return nil, nil
		}
		f.Imp.SetCurrentLabel(arc.Label())
		f.incr()
		return f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), f.fstReader)
	} else if idx == -1 {
		// Before first arc.
		return f.backtrackToFloorArc(arc, targetLabel, in)
	} else {
		// There is a floor arc; idx will be (-1 - (floor + 1)).
		f.fst.ReadArcByIndex(arc, in, -2-idx)

		f.pushLast()
		return nil, nil
	}
}

func (f *FSTEnumImp[T]) doSeekFloorList(arc *Arc[T], targetLabel int) (*Arc[T], error) {
	if arc.Label() == targetLabel {
		// Match -- recurse
		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if targetLabel == END_LABEL {
			return nil, nil
		}
		f.Imp.SetCurrentLabel(arc.Label())
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
				label, err := f.fst.readNextArcLabel(arc, f.fstReader)
				if err != nil {
					return nil, err
				}
				for !arc.IsLast() && label < targetLabel {
					f.fst.ReadNextArc(arc, f.fstReader)
					label, err = f.fst.readNextArcLabel(arc, f.fstReader)
					if err != nil {
						return nil, err
					}
				}
				f.pushLast()
				return nil, nil
			}
			f.upto--
			if f.upto == 0 {
				return nil, nil
			}
			targetLabel = f.Imp.GetTargetLabel()
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

// Seeks to exactly target term.
func (f *FSTEnumImp[T]) doSeekExact() (bool, error) {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	//System.out.println("FE: seek exact upto=" + upto);

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	f.rewindPrefix()

	//System.out.println("FE: after rewind upto=" + upto);
	arc := f.getArc(f.upto - 1)
	targetLabel := f.Imp.GetTargetLabel()

	fstReader := f.fst.GetBytesReader()

	for {
		//System.out.println("  cycle target=" + (targetLabel == -1 ? "-1" : (char) targetLabel));
		nextArc, err := f.fst.FindTargetArc(targetLabel, arc, f.getArc(f.upto), f.fstReader)
		if err != nil {
			return false, err
		}
		if nextArc == nil {
			// short circuit
			//upto--;
			//upto = 0;
			f.fst.ReadFirstTargetArc(arc, f.getArc(f.upto), fstReader)
			//System.out.println("  no match upto=" + upto);
			return false, nil
		}
		// Match -- recurse:
		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], nextArc.Output())
		if targetLabel == END_LABEL {
			//System.out.println("  return found; upto=" + upto + " output=" + output[upto] + " nextArc=" + nextArc.isLast());
			return true, nil
		}
		f.Imp.SetCurrentLabel(targetLabel)
		f.incr()
		targetLabel = f.Imp.GetTargetLabel()
		arc = nextArc
	}
}

func (f *FSTEnumImp[T]) incr() {
	f.upto++

	if len(f.arcs) <= f.upto {
		f.arcs = append(f.arcs, &Arc[T]{})
	}

	if len(f.output) <= f.upto {
		f.output = append(f.output, &Box[T]{})
	}
}

// Appends current arc, and then recurses from its target,
// appending first arc all the way to the final node
func (f *FSTEnumImp[T]) pushFirst() error {
	arc := f.arcs[f.upto]

	for {
		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if arc.Label() == END_LABEL {
			// Final node
			break
		}
		//System.out.println("  pushFirst label=" + (char) arc.label + " upto=" + upto + " output=" + fst.outputs.outputToString(output[upto]));
		f.Imp.SetCurrentLabel(arc.Label())
		f.incr()

		nextArc := f.getArc(f.upto)
		f.fst.ReadFirstTargetArc(arc, nextArc, f.fstReader)
		arc = nextArc
	}

	return nil
}

// Recurses from current arc, appending last arc all the
// way to the first final node
func (f *FSTEnumImp[T]) pushLast() error {
	arc := f.arcs[f.upto]

	for {
		f.Imp.SetCurrentLabel(arc.Label())
		f.output[f.upto] = f.fst.outputs.Add(f.output[f.upto-1], arc.Output())
		if arc.Label() == END_LABEL {
			break
		}
		f.incr()

		var err error
		arc, err = f.fst.readLastTargetArc(arc, f.getArc(f.upto), f.fstReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FSTEnumImp[T]) getArc(idx int) *Arc[T] {
	if f.arcs[idx] == nil {
		f.arcs[idx] = &Arc[T]{}
	}
	return f.arcs[idx]
}
