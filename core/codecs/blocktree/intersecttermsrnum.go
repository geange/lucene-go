package blocktree

import (
	"bytes"
	"context"
	"errors"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/array"
	"github.com/geange/lucene-go/core/util/automaton"
	"github.com/geange/lucene-go/core/util/fst"
)

var _ index.TermsEnum = &IntersectTermsEnum{}

type IntersectTermsEnum struct {
	*coreIndex.BaseTermsEnum

	in                store.IndexInput
	fstOutputs        fst.Outputs[[]byte]
	stack             []*IntersectTermsEnumFrame
	arcs              []*fst.Arc[[]byte]
	runAutomaton      *automaton.RunAutomaton
	automaton         *automaton.Automaton
	commonSuffix      []byte
	currentFrame      *IntersectTermsEnumFrame
	currentTransition *automaton.Transition
	term              []byte
	fstReader         fst.BytesReader
	fr                *FieldReader
	savedStartTerm    []byte
}

func (i *IntersectTermsEnum) Next(ctx context.Context) ([]byte, error) {
	term, err := i.next(ctx)
	if err != nil {
		if errors.Is(err, ErrNoMoreTerms) {
			return nil, nil
		}
		return nil, err
	}
	return term, nil
}

func (i *IntersectTermsEnum) next(ctx context.Context) ([]byte, error) {
	isSubBlock, err := i.popPushNext(ctx)
	if err != nil {
		return nil, err
	}

nextTerm:
	for {
		var state int
		var lastState int
		// NOTE: suffix == 0 can only happen on the first term in a block, when
		// there is a term exactly matching a prefix in the index.  If we
		// could somehow re-org the code so we only checked this case immediately
		// after pushing a frame...
		if i.currentFrame.suffix != 0 {
			suffixBytes := i.currentFrame.suffixBytes

			// This is the first byte of the suffix of the term we are now on:
			label := int(suffixBytes[i.currentFrame.startBytePos])

			if label < i.currentTransition.Min {
				// Common case: we are scanning terms in this block to "catch up" to
				// current transition in the automaton:
				minTrans := i.currentTransition.Min
				for i.currentFrame.nextEnt < i.currentFrame.entCount {
					isSubBlock, err = i.currentFrame.Next(ctx)
					if err != nil {
						return nil, err
					}
					if int(suffixBytes[i.currentFrame.startBytePos]) >= minTrans {
						continue nextTerm
					}
				}

				// End of frame:
				isSubBlock, err = i.popPushNext(ctx)
				if err != nil {
					return nil, err
				}
				continue nextTerm
			}
		} else {
			state = i.currentFrame.state
			lastState = i.currentFrame.lastState
		}

		if isSubBlock {
			// Match!  Recurse:
			i.copyTerm()
			currentFrame, err := i.pushFrame(ctx, state)
			if err != nil {
				return nil, err
			}
			i.currentFrame = currentFrame
			i.currentTransition = currentFrame.transition
			currentFrame.lastState = lastState
		} else if i.runAutomaton.IsAccept(state) {
			i.copyTerm()
			return i.term, nil
		} else {
			// This term is a prefix of a term accepted by the automaton, but is not itself accepted
		}
		isSubBlock, err = i.popPushNext(ctx)
		if err != nil {
			return nil, err
		}
	}
}

func (i *IntersectTermsEnum) copyTerm() {
	size := i.currentFrame.prefix + i.currentFrame.suffix
	array.Grow(i.term, size)
	i.term = i.term[:size]
}

var ErrNoMoreTerms = errors.New("no more terms")

func (i *IntersectTermsEnum) popPushNext(ctx context.Context) (bool, error) {
	// Pop finished frames
	for i.currentFrame.nextEnt == i.currentFrame.entCount {
		if !i.currentFrame.isLastInFloor {
			// Advance to next floor block
			if err := i.currentFrame.loadNextFloorBlock(ctx); err != nil {
				return false, err
			}
			break
		} else {
			if i.currentFrame.ord == 0 {
				return false, ErrNoMoreTerms
			}
			//lastFP := i.currentFrame.fpOrig
			i.currentFrame = i.stack[i.currentFrame.ord-1]
			i.currentTransition = i.currentFrame.transition
		}
	}

	return i.currentFrame.Next(ctx)
}

func (i *IntersectTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	return 0, errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	return false, errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	return errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) Term() ([]byte, error) {
	return i.term, nil
}

func (i *IntersectTermsEnum) Ord() (int64, error) {
	return 0, errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) DocFreq() (int, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return 0, err
	}
	return i.currentFrame.termState.GetDocFreq(), nil
}

func (i *IntersectTermsEnum) TotalTermFreq() (int64, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return 0, err
	}
	return int64(i.currentFrame.termState.GetTotalTermFreq()), nil
}

func (i *IntersectTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return nil, err
	}
	return i.fr.parent.postingsReader.Postings(
		context.Background(), i.fr.fieldInfo, i.currentFrame.termState, reuse, flags)
}

func (i *IntersectTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return nil, err
	}
	return i.fr.parent.postingsReader.Impacts(context.Background(),
		i.fr.fieldInfo, i.currentFrame.termState, flags)
}

// NOTE: specialized to only doing the first-time
// seek, but we could generalize it to allow
// arbitrary seekExact/Ceil.  Note that this is a
// seekFloor!
func (i *IntersectTermsEnum) seekToStartTerm(ctx context.Context, target []byte) error {
	if len(i.term) < len(target) {
		i.term = array.Grow(i.term, len(target))
	}

	for idx := 0; idx < len(target); idx++ {
		for {
			savNextEnt := i.currentFrame.nextEnt
			savePos := i.currentFrame.suffixesReader.GetPosition()
			saveLengthPos := i.currentFrame.suffixLengthsReader.GetPosition()
			saveStartBytePos := i.currentFrame.startBytePos
			saveSuffix := i.currentFrame.suffix
			saveLastSubFP := i.currentFrame.lastSubFP
			saveTermBlockOrd := i.currentFrame.termState.GetTermBlockOrd()

			//isSubBlock := i.currentFrame.Next()

			i.term = array.Grow(i.term, i.currentFrame.prefix+i.currentFrame.suffix)

			size := i.currentFrame.suffix

			copy(i.term[i.currentFrame.prefix:],
				i.currentFrame.suffixBytes[i.currentFrame.startBytePos:i.currentFrame.startBytePos+size])

			cmp := bytes.Compare(i.term, target)
			if cmp < 0 {
				if i.currentFrame.nextEnt == i.currentFrame.entCount {
					if !i.currentFrame.isLastInFloor {
						// Advance to next floor block
						err := i.currentFrame.loadNextFloorBlock(ctx)
						if err != nil {
							return err
						}
						continue
					} else {
						return nil
					}
				}
			} else if cmp == 0 {
				return nil
			} else {
				// Fallback to prior entry: the semantics of
				// this method is that the first call to
				// next() will return the term after the
				// requested term
				i.currentFrame.nextEnt = savNextEnt
				i.currentFrame.lastSubFP = saveLastSubFP
				i.currentFrame.startBytePos = saveStartBytePos
				i.currentFrame.suffix = saveSuffix
				i.currentFrame.suffixesReader.SetPosition(savePos)
				i.currentFrame.suffixLengthsReader.SetPosition(saveLengthPos)
				i.currentFrame.termState.SetTermBlockOrd(saveTermBlockOrd)

				copySize := i.currentFrame.suffix
				srcBytes := i.currentFrame.suffixBytes[i.currentFrame.startBytePos : i.currentFrame.startBytePos+copySize]
				dstBytes := i.term[i.currentFrame.prefix : i.currentFrame.prefix+copySize]
				copy(dstBytes, srcBytes)

				termSize := i.currentFrame.prefix + i.currentFrame.suffix
				i.term = i.term[:termSize]
				// If the last entry was a block we don't
				// need to bother recursing and pushing to
				// the last term under it because the first
				// next() will simply skip the frame anyway
				return nil
			}
		}
	}
	return nil
}

func (i *IntersectTermsEnum) TermState() (index.TermState, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return nil, err
	}
	return i.currentFrame.termState.Clone().(index.TermState), nil
}

func (i *IntersectTermsEnum) getFrame(ord int) (*IntersectTermsEnumFrame, error) {
	if ord >= len(i.stack) {
		frame, err := NewIntersectTermsEnumFrame(i, len(i.stack))
		if err != nil {
			return nil, err
		}
		i.stack = append(i.stack, frame)
	}
	return i.stack[ord], nil
}

func (i *IntersectTermsEnum) getArc(ord int) *fst.Arc[[]byte] {
	if ord >= len(i.arcs) {
		i.arcs = append(i.arcs, new(fst.Arc[[]byte]))
	}
	return i.arcs[ord]
}

func (i *IntersectTermsEnum) pushFrame(ctx context.Context, state int) (*IntersectTermsEnumFrame, error) {
	ord := 0
	if i.currentFrame != nil {
		ord = i.currentFrame.ord + 1
	}
	f, err := i.getFrame(ord)
	if err != nil {
		return nil, err
	}

	f.fp = i.currentFrame.lastSubFP
	f.fpOrig = i.currentFrame.lastSubFP
	f.prefix = i.currentFrame.prefix + i.currentFrame.suffix
	f.SetState(state)

	// Walk the arc through the index -- we only
	// "bother" with this so we can get the floor data
	// from the index and skip floor blocks when
	// possible:

	arc := i.currentFrame.arc
	idx := i.currentFrame.prefix
	output := fst.ByteSequenceOutput(i.currentFrame.outputPrefix)
	for idx < f.prefix {
		target := i.term[idx]
		// TODO: we could be more efficient for the next()
		// case by using current arc as starting point,
		// passed to findTargetArc
		arc, _, err = i.fr.index.FindTargetArc(ctx, int(target), i.fstReader, arc, i.getArc(1+idx))
		if err != nil {
			return nil, err
		}

		//assert arc != null;
		res, err := (output).Add(arc.Output())
		if err != nil {
			return nil, err
		}
		output = res.(fst.ByteSequenceOutput)
		idx++
	}

	f.arc = arc
	f.outputPrefix = output

	frameIndexData, err := output.Add(arc.NextFinalOutput())
	if err != nil {
		return nil, err
	}

	if err := f.load(ctx, frameIndexData.(fst.ByteSequenceOutput)); err != nil {
		return nil, err
	}
	return f, nil
}
