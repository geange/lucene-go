package fst

import (
	"context"
	"slices"
)

// Enum Can next() and advance() through the terms in an FST
type Enum struct {
	fst          *FST
	arcs         []*Arc   //
	output       []Output // outputs are cumulative
	noOutput     Output
	fstReader    BytesReader
	upto         int
	targetLength int
}

type LabelManager interface {
	GetTargetLabel(upto int) int
	GetCurrentLabel(upto int) int
	SetCurrentLabel(label int) error
	Grow()
}

func NewEnum(fst *FST) (*Enum, error) {
	reader, err := fst.GetBytesReader()
	if err != nil {
		return nil, err
	}

	noOutput := fst.manager.EmptyOutput()

	enum := &Enum{
		fst:       fst,
		fstReader: reader,
		noOutput:  noOutput,
		arcs:      make([]*Arc, 10),
		output:    make([]Output, 10),
	}

	enum.output[0] = noOutput

	for i := range enum.arcs {
		enum.arcs[i] = &Arc{}
	}

	if _, err := fst.GetFirstArc(enum.getArc(0)); err != nil {
		return nil, err
	}

	return enum, nil
}

func (r *Enum) GetUpTo() int {
	return r.upto
}

func (r *Enum) GetOutput(idx int) Output {
	return r.output[idx]
}

func (r *Enum) SetTargetLength(size int) {
	r.targetLength = size
}

// Rewinds enum state to match the shared prefix between current term and target term
// 倒回枚举状态，以匹配当前term和目标term之间的共享前缀
func (r *Enum) rewindPrefix(ctx context.Context, manager LabelManager) error {
	if r.upto == 0 {
		r.upto = 1
		if _, err := r.fst.ReadFirstTargetArc(ctx, r.fstReader, r.getArc(0), r.getArc(1)); err != nil {
			return err
		}
	}

	currentLimit := r.upto
	r.upto = 1

	targetLengthPlus := r.targetLength + 1

	for r.upto < currentLimit && r.upto < targetLengthPlus {
		label1 := manager.GetCurrentLabel(r.upto)
		label2 := manager.GetTargetLabel(r.upto)

		cmp := label1 - label2
		if cmp < 0 {
			// seek forward
			// 向前搜素，停止计算
			break
		}

		if cmp > 0 {
			// seek backwards -- reset this arc to the first arc
			// 向后搜索
			follow := r.arcs[r.upto-1]
			arc := r.getArc(r.upto)
			if _, err := r.fst.ReadFirstTargetArc(ctx, r.fstReader, follow, arc); err != nil {
				return err
			}
			break
		}

		r.upto++
	}

	return nil
}

func (r *Enum) DoNext(ctx context.Context, lm LabelManager) error {
	if r.upto == 0 {
		r.upto = 1
		follow := r.getArc(0)
		arc := r.getArc(1)
		if _, err := r.fst.ReadFirstTargetArc(ctx, r.fstReader, follow, arc); err != nil {
			return err
		}
		return r.pushFirst(ctx, lm)
	}

	// pop
	for r.arcs[r.upto].IsLast() {
		r.upto--
		if r.upto == 0 {
			return nil
		}
	}
	if _, err := r.fst.ReadNextArc(ctx, r.arcs[r.upto], r.fstReader); err != nil {
		return err
	}
	return r.pushFirst(ctx, lm)
}

// DoSeekCeil
// Seeks to smallest term that's >= target.
func (r *Enum) DoSeekCeil(ctx context.Context, lm LabelManager) error {

	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	if err := r.rewindPrefix(ctx, lm); err != nil {
		return err
	}

	arc := r.getArc(r.upto)

	var err error

	for arc != nil {
		targetLabel := lm.GetTargetLabel(r.upto)

		if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
			// Arcs are in an array
			in, err := r.fst.GetBytesReader()
			if err != nil {
				return err
			}
			if arc.NodeFlags() == ArcsForDirectAddressing {
				arc, err = r.doSeekCeilArrayDirectAddressing(ctx, targetLabel, in, lm, arc)
				if err != nil {
					return err
				}
			} else {
				// assert arc.nodeFlags() == Fst.ARCS_FOR_BINARY_SEARCH;
				arc, err = r.doSeekCeilArrayPacked(ctx, targetLabel, in, lm, arc)
				if err != nil {
					return err
				}
			}
		} else {
			arc, err = r.doSeekCeilList(ctx, arc, lm, targetLabel)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DoSeekFloor
// Seeks to largest term that's <= target.
func (r *Enum) DoSeekFloor(ctx context.Context, lm LabelManager) error {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save CPU by starting at the end of the shared prefix
	// b/w our current term & the target:
	// 从当前术语和目标的共享前缀末尾开始节省 CPU
	if err := r.rewindPrefix(ctx, lm); err != nil {
		return err
	}

	arc := r.getArc(r.upto)

	var err error

	// Now scan forward, matching the new suffix of the target
	for arc != nil {
		targetLabel := lm.GetTargetLabel(r.upto)

		if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
			// Arcs are in an array
			in, err := r.fst.GetBytesReader()
			if err != nil {
				return err
			}
			if arc.NodeFlags() == ArcsForDirectAddressing {
				arc, err = r.doSeekFloorArrayDirectAddressing(ctx, targetLabel, in, lm, arc)
				if err != nil {
					return err
				}
			} else {
				arc, err = r.doSeekFloorArrayPacked(ctx, targetLabel, in, lm, arc)
				if err != nil {
					return err
				}
			}
		} else {
			arc, err = r.doSeekFloorList(ctx, arc, lm, targetLabel)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DoSeekExact Seeks to exactly target term.
func (r *Enum) DoSeekExact(ctx context.Context, lm LabelManager) (bool, error) {
	// TODO: possibly caller could/should provide common
	// prefix length?  ie this work may be redundant if
	// caller is in fact intersecting against its own
	// automaton

	// Save time by starting at the end of the shared prefix
	// b/w our current term & the target:
	if err := r.rewindPrefix(ctx, lm); err != nil {
		return false, err
	}

	arc := r.getArc(r.upto - 1)
	targetLabel := lm.GetTargetLabel(r.upto)

	fstReader, err := r.fst.GetBytesReader()
	if err != nil {
		return false, err
	}

	for {
		nextArc, found, err := r.fst.FindTargetArc(ctx, targetLabel, fstReader, arc, r.getArc(r.upto))
		if err != nil {
			return false, err
		}

		if !found {
			if _, err := r.fst.ReadFirstTargetArc(ctx, fstReader, arc, r.getArc(r.upto)); err != nil {
				return false, err
			}
			return false, nil
		}

		// Match -- recurse:
		r.output[r.upto], err = r.output[r.upto-1].Add(nextArc.Output())
		if err != nil {
			return false, err
		}

		if targetLabel == END_LABEL {
			return true, nil
		}

		if err = lm.SetCurrentLabel(targetLabel); err != nil {
			return false, err
		}

		r.incr(lm)

		targetLabel = lm.GetTargetLabel(r.upto)

		arc = nextArc
	}
}

func (r *Enum) doSeekCeilArrayDirectAddressing(ctx context.Context, targetLabel int, in BytesReader, lm LabelManager, arc *Arc) (*Arc, error) {

	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex >= arc.NumArcs() {
		// Target is beyond the last arc, out of label range.
		// Dead end (target is after the last arc);
		// rollback to last fork then push
		r.upto--
		for {
			if r.upto == 0 {
				return nil, nil
			}

			prevArc := r.getArc(r.upto)
			if !prevArc.IsLast() {
				if _, err := r.fst.ReadNextArc(ctx, prevArc, r.fstReader); err != nil {
					return nil, err
				}

				if err := r.pushFirst(ctx, lm); err != nil {
					return nil, err
				}
				return nil, nil
			}
			r.upto--
		}
	}

	if targetIndex < 0 {
		targetIndex = -1
	} else if ok, err := IsBitSet(ctx, targetIndex, arc, in); ok && err == nil {
		if _, err := r.fst.ReadArcByDirectAddressing(ctx, in, targetIndex, arc); err != nil {
			return nil, err
		}
		// found -- copy pasta from below
		r.output[r.upto], err = r.output[r.upto-1].Add(arc.Output())
		if targetLabel == END_LABEL {
			return nil, nil
		}
		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return nil, err
		}
		r.incr(lm)
		return r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
	}
	// Not found, return the next arc (ceil).
	ceilIndex, err := NextBitSet(ctx, targetIndex, arc, in)
	if err != nil {
		return nil, err
	}

	if _, err := r.fst.ReadArcByDirectAddressing(ctx, in, ceilIndex, arc); err != nil {
		return nil, err
	}

	if err := r.pushFirst(ctx, lm); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Enum) doSeekCeilArrayPacked(ctx context.Context, targetLabel int, in BytesReader, lm LabelManager, arc *Arc) (*Arc, error) {
	// The array is packed -- use binary search to find the target.
	idx, err := binarySearch(ctx, r.fst, arc, targetLabel)
	if err != nil {
		return nil, err
	}

	if (idx) >= 0 {
		// Match
		if _, err := r.fst.ReadArcByIndex(ctx, in, idx, arc); err != nil {
			return nil, err
		}
		//assert arc.arcIdx() == idx;
		//assert arc.label() == targetLabel: "arc.label=" + arc.label() + " vs targetLabel=" + targetLabel + " mid=" + idx;
		r.output[r.upto], err = r.output[r.upto-1].Add(arc.Output())
		if targetLabel == END_LABEL {
			return nil, err
		}
		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return nil, err
		}
		r.incr(lm)
		return r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
	}

	idx = -1 - idx
	if idx == arc.NumArcs() {
		// Dead end
		if _, err := r.fst.ReadArcByIndex(ctx, in, idx-1, arc); err != nil {
			return nil, err
		}
		//assert arc.isLast();
		// Dead end (target is after the last arc);
		// rollback to last fork then push
		r.upto--
		for {
			if r.upto == 0 {
				return nil, nil
			}
			prevArc := r.getArc(r.upto)
			if !prevArc.IsLast() {
				if _, err := r.fst.ReadNextArc(ctx, prevArc, r.fstReader); err != nil {
					return nil, err
				}
				if err = r.pushFirst(ctx, lm); err != nil {
					return nil, err
				}
				return nil, nil
			}
			r.upto--
		}
	} else {
		// ceiling - arc with least higher label
		if _, err := r.fst.ReadArcByIndex(ctx, in, idx, arc); err != nil {
			return nil, err
		}
		//assert arc.label() > targetLabel;
		if err = r.pushFirst(ctx, lm); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (r *Enum) doSeekCeilList(ctx context.Context, arc *Arc, lm LabelManager, targetLabel int) (*Arc, error) {
	// Arcs are not array'd -- must do linear scan:
	if arc.Label() == targetLabel {
		// recurse
		output, err := r.output[r.upto-1].Add(arc.Output())
		if err != nil {
			return nil, err
		}
		r.output[r.upto] = output

		if targetLabel == END_LABEL {
			return nil, nil
		}

		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return nil, err
		}
		r.incr(lm)
		return r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
	}

	if arc.Label() > targetLabel {
		if err := r.pushFirst(ctx, lm); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if arc.IsLast() {
		// Dead end (target is after the last arc);
		// rollback to last fork then push
		r.upto--
		for {
			if r.upto == 0 {
				return nil, nil
			}
			prevArc := r.getArc(r.upto)
			if !prevArc.IsLast() {
				if _, err := r.fst.ReadNextArc(ctx, prevArc, r.fstReader); err != nil {
					return nil, err
				}

				if err := r.pushFirst(ctx, lm); err != nil {
					return nil, err
				}
				return nil, nil
			}
			r.upto--
		}
	}

	// keep scanning
	if _, err := r.fst.ReadNextArc(ctx, arc, r.fstReader); err != nil {
		return nil, err
	}
	return arc, nil
}

func (r *Enum) doSeekFloorArrayDirectAddressing(ctx context.Context, targetLabel int, in BytesReader, lm LabelManager, arc *Arc) (*Arc, error) {
	// The array is addressed directly by label, with presence bits to compute the actual arc offset.

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex < 0 {
		// Before first arc.
		return r.backtrackToFloorArc(ctx, targetLabel, in, lm, arc)
	}

	if targetIndex >= arc.NumArcs() {
		// After last arc.
		if _, err := r.fst.ReadLastArcByDirectAddressing(ctx, arc, in); err != nil {
			return nil, err
		}
		if err := r.pushLast(ctx, lm); err != nil {
			return nil, err
		}
		return nil, nil
	}

	var err error
	// Within label range.
	if ok, _ := IsBitSet(ctx, targetIndex, arc, in); ok {
		if _, err := r.fst.ReadArcByDirectAddressing(ctx, in, targetIndex, arc); err != nil {
			return nil, err
		}
		// found -- copy pasta from below
		output, err := r.output[r.upto-1].Add(arc.Output())
		if err != nil {
			return nil, err
		}
		r.output[r.upto] = output

		if targetLabel == END_LABEL {
			return nil, nil
		}

		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return nil, err
		}
		r.incr(lm)
		return r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
	}
	// Scan backwards to find a floor arc.
	floorIndex, err := PreviousBitSet(targetIndex, arc, in)
	if err != nil {
		return nil, err
	}

	if _, err := r.fst.ReadArcByDirectAddressing(ctx, in, floorIndex, arc); err != nil {
		return nil, err
	}

	if err := r.pushLast(ctx, lm); err != nil {
		return nil, err
	}
	return nil, nil
}

// Backtracks until it finds a node which first arc is before our target label.`
// Then on the node, finds the arc just before the targetLabel.
// return null to continue the seek floor recursion loop.
func (r *Enum) backtrackToFloorArc(ctx context.Context, targetLabel int, in BytesReader, lm LabelManager, arc *Arc) (*Arc, error) {
	for {
		// First, walk backwards until we find a node which first arc is before our target label.
		follow := r.getArc(r.upto - 1)
		if _, err := r.fst.ReadFirstTargetArc(ctx, r.fstReader, follow, arc); err != nil {
			return nil, err
		}
		if arc.Label() < targetLabel {
			// Then on this node, find the arc just before the targetLabel.
			if !arc.IsLast() {
				if arc.BytesPerArc() != 0 && arc.Label() != END_LABEL {
					if arc.NodeFlags() == ArcsForBinarySearch {
						if err := r.findNextFloorArcBinarySearch(ctx, arc, targetLabel, in); err != nil {
							return nil, err
						}
					} else {
						if err := r.findNextFloorArcDirectAddressing(ctx, arc, targetLabel, in); err != nil {
							return nil, err
						}
					}
				} else {
					for {
						if !arc.IsLast() {
							if n, _ := r.fst.readNextArcLabel(ctx, arc, in); n < targetLabel {
								if _, err := r.fst.ReadNextArc(ctx, arc, r.fstReader); err != nil {
									return nil, err
								}
								continue
							}
						}
						break
					}
				}
			}

			if err := r.pushLast(ctx, lm); err != nil {
				return nil, err
			}
			return nil, nil
		}

		r.upto--
		if r.upto == 0 {
			return nil, nil
		}

		targetLabel = lm.GetTargetLabel(r.upto)
		arc = r.getArc(r.upto)
	}
}

// Finds and reads an arc on the current node which label is strictly less than the given label.
// Skips the first arc, finds next floor arc; or none if the floor arc is the first arc itself
// (in this case it has already been read).
// Precondition: the given arc is the first arc of the node.
func (r *Enum) findNextFloorArcDirectAddressing(ctx context.Context, arc *Arc, targetLabel int, in BytesReader) error {
	if arc.NumArcs() <= 1 {
		return nil
	}

	targetIndex := targetLabel - arc.FirstLabel()
	if targetIndex >= arc.NumArcs() {
		// Beyond last arc. Take last arc.
		if _, err := r.fst.ReadLastArcByDirectAddressing(ctx, arc, in); err != nil {
			return err
		}
		return nil
	}

	// Take the preceding arc, even if the target is present.
	floorIndex, err := PreviousBitSet(targetIndex, arc, in)
	if err != nil {
		return err
	}
	if floorIndex > 0 {
		if _, err := r.fst.ReadArcByDirectAddressing(ctx, in, floorIndex, arc); err != nil {
			return err
		}
	}
	return nil
}

// Same as findNextFloorArcDirectAddressing for binary search node.
func (r *Enum) findNextFloorArcBinarySearch(ctx context.Context, arc *Arc, targetLabel int, in BytesReader) error {
	if arc.NumArcs() > 1 {
		idx, err := binarySearch(ctx, r.fst, arc, targetLabel)
		if err != nil {
			return err
		}

		if idx > 1 {
			if _, err := r.fst.ReadArcByIndex(ctx, in, idx-1, arc); err != nil {
				return err
			}
		} else if idx < -2 {
			if _, err := r.fst.ReadArcByIndex(ctx, in, -2-idx, arc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Enum) doSeekFloorArrayPacked(ctx context.Context, targetLabel int, in BytesReader, lm LabelManager, arc *Arc) (*Arc, error) {
	// Arcs are fixed array -- use binary search to find the target.
	idx, err := binarySearch(ctx, r.fst, arc, targetLabel)
	if err != nil {
		return nil, err
	}

	switch {
	case idx >= 0:
		// Match -- recurse
		if _, err := r.fst.ReadArcByIndex(ctx, in, idx, arc); err != nil {
			return nil, err
		}

		output, err := r.output[r.upto-1].Add(arc.Output())
		if err != nil {
			return nil, err
		}
		r.output[r.upto] = output

		if targetLabel == END_LABEL {
			return nil, nil
		}

		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return nil, err
		}

		r.incr(lm)
		return r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
	case idx == -1:
		// Before first arc.
		return r.backtrackToFloorArc(ctx, targetLabel, in, lm, arc)
	default:
		// There is a floor arc; idx will be (-1 - (floor + 1)).
		if _, err := r.fst.ReadArcByIndex(ctx, in, -2-idx, arc); err != nil {
			return nil, err
		}
		//assert arc.isLast() || fst.readNextArcLabel(arc, in) > targetLabel;
		//assert arc.label() < targetLabel: "arc.label=" + arc.label() + " vs targetLabel=" + targetLabel;
		if err := r.pushLast(ctx, lm); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (r *Enum) doSeekFloorList(ctx context.Context, arc *Arc, lm LabelManager, targetLabel int) (*Arc, error) {
	if arc.Label() == targetLabel {
		// Match -- recurse
		output, err := r.output[r.upto-1].Add(arc.Output())
		if err != nil {
			return nil, err
		}
		r.output[r.upto] = output

		if targetLabel == END_LABEL {
			return nil, nil
		}

		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return nil, err
		}

		r.incr(lm)
		return r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
	}

	if arc.Label() > targetLabel {
		// TODO: if each arc could somehow read the arc just
		// before, we can save this re-scan.  The ceil case
		// doesn't need this because it reads the next arc
		// instead:
		for {
			// First, walk backwards until we find a first arc
			// that's before our target label:
			if _, err := r.fst.ReadFirstTargetArc(ctx, r.fstReader, r.getArc(r.upto-1), arc); err != nil {
				return nil, err
			}
			if arc.Label() < targetLabel {
				// Then, scan forwards to the arc just before
				// the targetLabel:
				for {
					if !arc.IsLast() {
						if n, _ := r.fst.readNextArcLabel(ctx, arc, r.fstReader); n < targetLabel {
							if _, err := r.fst.ReadNextArc(ctx, arc, r.fstReader); err != nil {
								return nil, err
							}
							continue
						}
					}
					break
				}

				if err := r.pushLast(ctx, lm); err != nil {
					return nil, err
				}
				return nil, nil
			}
			r.upto--
			if r.upto == 0 {
				return nil, nil
			}
			targetLabel = lm.GetTargetLabel(r.upto)
			arc = r.getArc(r.upto)
		}
	}

	if !arc.IsLast() {
		if n, _ := r.fst.readNextArcLabel(ctx, arc, r.fstReader); n > targetLabel {
			if err := r.pushLast(ctx, lm); err != nil {
				return nil, err
			}
			return nil, nil
		}
		// keep scanning
		return r.fst.ReadNextArc(ctx, arc, r.fstReader)
	}

	if err := r.pushLast(ctx, lm); err != nil {
		return nil, err
	}
	return nil, nil
}

// Appends current arc, and then recurses from its target,
// appending first arc all the way to the final node
func (r *Enum) pushFirst(ctx context.Context, lm LabelManager) error {

	arc := r.arcs[r.upto]

	for {
		output, err := r.output[r.upto-1].Add(arc.Output())
		if err != nil {
			return err
		}
		r.output[r.upto] = output

		if arc.Label() == END_LABEL {
			// Final node
			break
		}

		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return err
		}
		r.incr(lm)

		nextArc := r.getArc(r.upto)
		if _, err := r.fst.ReadFirstTargetArc(ctx, r.fstReader, arc, nextArc); err != nil {
			return err
		}
		arc = nextArc
	}
	return nil
}

// Recurse from current arc, appending last arc all the
// way to the first final node
func (r *Enum) pushLast(ctx context.Context, lm LabelManager) error {
	arc := r.arcs[r.upto]

	for {
		if err := lm.SetCurrentLabel(arc.Label()); err != nil {
			return err
		}

		output, err := r.output[r.upto-1].Add(arc.Output())
		if err != nil {
			return err
		}
		r.output[r.upto] = output

		if arc.Label() == END_LABEL {
			// Final node
			break
		}

		r.incr(lm)

		arc, err = r.fst.readLastTargetArc(ctx, r.fstReader, arc, r.getArc(r.upto))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Enum) getArc(idx int) *Arc {
	if r.arcs[idx] == nil {
		r.arcs[idx] = &Arc{}
	}
	return r.arcs[idx]
}

func (r *Enum) incr(lm LabelManager) {
	r.upto++
	lm.Grow()
	r.arcs = slices.Grow(r.arcs, r.upto+1)
	r.output = slices.Grow(r.output, r.upto+1)
}
