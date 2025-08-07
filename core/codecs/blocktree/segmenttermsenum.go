package blocktree

import (
	"bytes"
	"context"
	"errors"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
	"golang.org/x/exp/slog"
)

var _ index.TermsEnum = &SegmentTermsEnum{}

func NewSegmentTermsEnum(fr *FieldReader) (*SegmentTermsEnum, error) {
	this := &SegmentTermsEnum{
		scratchReader: store.NewByteArrayDataInput(nil),
		arcs:          make([]*fst.Arc[[]byte], 1),
		fr:            fr,
		stack:         make([]*SegmentTermsEnumFrame, 0),
	}

	// Used to hold seek by TermState, or cached seek
	staticFrame, err := NewSegmentTermsEnumFrame(this, -1)
	if err != nil {
		return nil, err
	}
	this.staticFrame = staticFrame

	if fr.index == nil {
		this.fstReader = nil
	} else {
		fstReader, err := fr.index.GetBytesReader()
		if err != nil {
			return nil, err
		}
		this.fstReader = fstReader
	}

	// Init w/ root block; don't use index since it may
	// not (and need not) have been loaded
	for i := range this.arcs {
		this.arcs[i] = new(fst.Arc[[]byte])
	}

	this.currentFrame = staticFrame
	//var arc *fst.Arc[[]byte]
	if fr.index != nil {
		_, err = fr.index.GetFirstArc(this.arcs[0])
		if err != nil {
			return nil, err
		}
		// Empty string prefix must have an output in the index!
		//assert arc.isFinal();
	} else {
		//arc = nil
	}
	//currentFrame = pushFrame(arc, rootCode, 0);
	//currentFrame.loadBlock();
	this.validIndexPrefix = 0
	return this, nil
}

type SegmentTermsEnum struct {
	*coreIndex.BaseTermsEnum

	// Lazy init:
	in store.IndexInput

	stack        []*SegmentTermsEnumFrame
	staticFrame  *SegmentTermsEnumFrame
	currentFrame *SegmentTermsEnumFrame
	termExists   bool
	fr           *FieldReader

	targetBeforeCurrentLength int

	//static boolean DEBUG = BlockTreeTermsWriter.DEBUG;

	scratchReader *store.ByteArrayDataInput

	// What prefix of the current term was present in the index; when we only next() through the index, this stays at 0.  It's only set when
	// we seekCeil/Exact:
	validIndexPrefix int

	// assert only:
	eof bool

	term      []byte
	fstReader fst.BytesReader

	arcs []*fst.Arc[[]byte]
}

func (s *SegmentTermsEnum) Next(ctx context.Context) ([]byte, error) {
	var err error

	if s.in == nil {

		// Fresh TermsEnum; seek to first term:
		var arc *fst.Arc[[]byte]

		if s.fr.index != nil {
			arc, err = s.fr.index.GetFirstArc(s.arcs[0])
			// Empty string prefix must have an output in the index!
			//assert arc.isFinal();
		} else {
			arc = nil
		}

		s.currentFrame, err = s.pushFrameSeekTo(ctx, arc, s.fr.rootCode, 0)
		if err != nil {
			return nil, err
		}
		if err := s.currentFrame.loadBlock(ctx); err != nil {
			return nil, err
		}
	}

	s.targetBeforeCurrentLength = s.currentFrame.ord

	//assert !eof;
	// if (DEBUG) {
	//   System.out.println("\nBTTR.next seg=" + fr.parent.segment + " term=" + brToString(term) + " termExists?=" + termExists + " field=" + fr.fieldInfo.name + " termBlockOrd=" + currentFrame.state.termBlockOrd + " validIndexPrefix=" + validIndexPrefix);
	//   printSeekState(System.out);
	// }

	if s.currentFrame == s.staticFrame {
		// If seek was previously called and the term was
		// cached, or seek(TermState) was called, usually
		// caller is just going to pull a D/&PEnum or get
		// docFreq, etc.  But, if they then call next(),
		// this method catches up all internal state so next()
		// works properly:
		//if (DEBUG) System.out.println("  re-seek to pending term=" + term.utf8ToString() + " " + term);
		if _, err := s.SeekExact(ctx, s.term); err != nil {
			return nil, err
		}
		//assert result;
	}

	// Pop finished blocks
	for s.currentFrame.nextEnt == s.currentFrame.entCount {
		if !s.currentFrame.isLastInFloor {
			// Advance to next floor block
			if err := s.currentFrame.loadNextFloorBlock(ctx); err != nil {
				return nil, err
			}
			break
		} else {
			//if (DEBUG) System.out.println("  pop frame");
			if s.currentFrame.ord == 0 {
				//if (DEBUG) System.out.println("  return null");
				//assert setEOF();
				s.term = s.term[:0]
				s.validIndexPrefix = 0
				if err := s.currentFrame.rewind(ctx); err != nil {
					return nil, err
				}
				s.termExists = false
				return nil, nil
			}
			lastFP := s.currentFrame.fpOrig
			s.currentFrame = s.stack[s.currentFrame.ord-1]

			if s.currentFrame.nextEnt == -1 || s.currentFrame.lastSubFP != lastFP {
				// We popped into a frame that's not loaded
				// yet or not scan'd to the right entry
				if err := s.currentFrame.scanToFloorFrame(ctx, s.term); err != nil {
					return nil, err
				}
				if err := s.currentFrame.loadBlock(ctx); err != nil {
					return nil, err
				}
				if err := s.currentFrame.scanToSubBlock(ctx, lastFP); err != nil {
					return nil, err
				}
			}

			// Note that the seek state (last seek) has been
			// invalidated beyond this depth
			s.validIndexPrefix = min(s.validIndexPrefix, s.currentFrame.prefix)
			//if (DEBUG) {
			//System.out.println("  reset validIndexPrefix=" + validIndexPrefix);
			//}
		}
	}

	for {
		if ok, _ := s.currentFrame.next(ctx); ok {
			// Push to new block:
			//if (DEBUG) System.out.println("  push frame");
			s.currentFrame, err = s.pushFrame(ctx, nil, s.currentFrame.lastSubFP, len(s.term))
			// This is a "next" frame -- even if it's
			// floor'd we must pretend it isn't so we don't
			// try to scan to the right floor frame:
			if err := s.currentFrame.loadBlock(ctx); err != nil {
				return nil, err
			}
		} else {
			//if (DEBUG) System.out.println("  return term=" + brToString(term) + " currentFrame.ord=" + currentFrame.ord);
			return s.term, nil
		}
	}
}

func ensureSize(a []byte, size int) []byte {
	if len(a) < size {
		newArray := make([]byte, size)
		copy(newArray, a)
		return newArray
	}
	return a[:size]
}

func (s *SegmentTermsEnum) SeekCeil(ctx context.Context, target []byte) (index.SeekStatus, error) {

	if s.fr.index == nil {
		return index.SEEK_STATUS_UNDEFINED, errors.New("terms index was not loaded")
	}

	s.term = ensureSize(s.term, len(target))

	//s.term.grow(1 + target.length)

	//assert clearEOF();

	// if (DEBUG) {
	//   System.out.println("\nBTTR.seekCeil seg=" + fr.parent.segment + " target=" + fr.fieldInfo.name + ":" + brToString(target) + " " + target + " current=" + brToString(term) + " (exists?=" + termExists + ") validIndexPrefix=  " + validIndexPrefix);
	//   printSeekState(System.out);
	// }

	var arc *fst.Arc[[]byte]
	var targetUpto int
	var output []byte

	s.targetBeforeCurrentLength = s.currentFrame.ord

	if s.currentFrame != s.staticFrame {

		// We are already seek'd; find the common
		// prefix of new seek term vs current term and
		// re-use the corresponding seek state.  For
		// example, if app first seeks to foobar, then
		// seeks to foobaz, we can re-use the seek state
		// for the first 5 bytes.

		//if (DEBUG) {
		//System.out.println("  re-use current seek state validIndexPrefix=" + validIndexPrefix);
		//}

		arc = s.arcs[0]
		//assert arc.isFinal();
		output = arc.Output()
		targetUpto = 0
		var err error

		lastFrame := s.stack[0]
		//assert validIndexPrefix <= term.length();

		targetLimit := min(len(target), s.validIndexPrefix)

		cmp := 0

		// TODO: we should write our vLong backwards (MSB
		// first) to get better sharing from the FST

		// First compare up to valid seek frames:
		for targetUpto < targetLimit {
			cmp = int(s.term[targetUpto] - target[targetUpto])
			//if (DEBUG) {
			//System.out.println("    cycle targetUpto=" + targetUpto + " (vs limit=" + targetLimit + ") cmp=" + cmp + " (targetLabel=" + (char) (target.bytes[target.offset + targetUpto]) + " vs termLabel=" + (char) (term.byteAt(targetUpto)) + ")"   + " arc.output=" + arc.output + " output=" + output);
			//}
			if cmp != 0 {
				break
			}
			arc = s.arcs[1+targetUpto]
			//assert arc.label() == (target.bytes[target.offset + targetUpto] & 0xFF): "arc.label=" + (char) arc.label() + " targetLabel=" + (char) (target.bytes[target.offset + targetUpto] & 0xFF);
			// TODO: we could save the outputs in local
			// byte[][] instead of making new objs ever
			// seek; but, often the FST doesn't have any
			// shared bytes (but this could change if we
			// reverse vLong byte order)
			if len(arc.Output()) != 0 {
				output, err = FST_OUTPUTS.Add(output, arc.Output())
				if err != nil {
					return index.SEEK_STATUS_UNDEFINED, err
				}
			}
			if arc.IsFinal() {
				lastFrame = s.stack[1+lastFrame.ord]
			}
			targetUpto++
		}

		if cmp == 0 {
			targetUptoMid := targetUpto
			// Second compare the rest of the term, but
			// don't save arc/output/frame:
			targetLimit2 := min(len(target), len(s.term))
			for targetUpto < targetLimit2 {
				cmp = int(s.term[targetUpto] - target[targetUpto])
				//if (DEBUG) {
				//System.out.println("    cycle2 targetUpto=" + targetUpto + " (vs limit=" + targetLimit + ") cmp=" + cmp + " (targetLabel=" + (char) (target.bytes[target.offset + targetUpto]) + " vs termLabel=" + (char) (term.byteAt(targetUpto)) + ")");
				//}
				if cmp != 0 {
					break
				}
				targetUpto++
			}

			if cmp == 0 {
				cmp = len(s.term) - len(target)
			}
			targetUpto = targetUptoMid
		}

		if cmp < 0 {
			// Common case: target term is after current
			// term, ie, app is seeking multiple terms
			// in sorted order
			//if (DEBUG) {
			//System.out.println("  target is after current (shares prefixLen=" + targetUpto + "); clear frame.scanned ord=" + lastFrame.ord);
			//}
			s.currentFrame = lastFrame

		} else if cmp > 0 {
			// Uncommon case: target term
			// is before current term; this means we can
			// keep the currentFrame but we must rewind it
			// (so we scan from the start)
			s.targetBeforeCurrentLength = 0
			//if (DEBUG) {
			//System.out.println("  target is before current (shares prefixLen=" + targetUpto + "); rewind frame ord=" + lastFrame.ord);
			//}
			s.currentFrame = lastFrame
			if err := s.currentFrame.rewind(ctx); err != nil {
				return index.SEEK_STATUS_UNDEFINED, err
			}
		} else {
			// Target is exactly the same as current term
			//assert term.length() == target.length;
			if s.termExists {
				//if (DEBUG) {
				//System.out.println("  target is same as current; return FOUND");
				//}
				return index.SEEK_STATUS_END, nil
			} else {
				//if (DEBUG) {
				//System.out.println("  target is same as current but term doesn't exist");
				//}
			}
		}

	} else {
		var err error
		s.targetBeforeCurrentLength = -1
		arc, err = s.fr.index.GetFirstArc(s.arcs[0])
		if err != nil {
			return index.SEEK_STATUS_UNDEFINED, err
		}

		// Empty string prefix must have an output (block) in the index!
		//assert arc.isFinal();
		//assert arc.output() != null;

		//if (DEBUG) {
		//System.out.println("    no seek state; push root frame");
		//}

		output = arc.Output()

		s.currentFrame = s.staticFrame

		//term.length = 0;
		targetUpto = 0
		frameData, err := FST_OUTPUTS.Add(output, arc.NextFinalOutput())
		if err != nil {
			return index.SEEK_STATUS_UNDEFINED, err
		}
		s.currentFrame, err = s.pushFrameSeekTo(ctx, arc, frameData, 0)
	}

	//if (DEBUG) {
	//System.out.println("  start index loop targetUpto=" + targetUpto + " output=" + output + " currentFrame.ord+1=" + currentFrame.ord + " targetBeforeCurrentLength=" + targetBeforeCurrentLength);
	//}

	// We are done sharing the common prefix with the incoming target and where we are currently seek'd; now continue walking the index:
	for targetUpto < len(target) {

		targetLabel := target[targetUpto]

		nextArc, found, err := s.fr.index.FindTargetArc(ctx, int(targetLabel), s.fstReader, arc, s.getArc(1+targetUpto))
		if err != nil {
			return index.SEEK_STATUS_UNDEFINED, err
		}

		if found {

			// Index is exhausted
			// if (DEBUG) {
			//   System.out.println("    index: index exhausted label=" + ((char) targetLabel) + " " + targetLabel);
			// }

			s.validIndexPrefix = s.currentFrame.prefix
			//validIndexPrefix = targetUpto;

			if err := s.currentFrame.scanToFloorFrame(ctx, target); err != nil {
				return index.SEEK_STATUS_UNDEFINED, err
			}

			if err := s.currentFrame.loadBlock(ctx); err != nil {
				return index.SEEK_STATUS_UNDEFINED, err
			}

			//if (DEBUG) System.out.println("  now scanToTerm");
			result, err := s.currentFrame.scanToTerm(ctx, target, false)
			if err != nil {
				return index.SEEK_STATUS_UNDEFINED, err
			}
			if result == index.SEEK_STATUS_END {
				copy(s.term, target)
				//s.term.copyBytes(target)
				s.termExists = false

				if _, err := s.Next(ctx); err == nil {
					//if (DEBUG) {
					//System.out.println("  return NOT_FOUND term=" + brToString(term));
					//}
					return index.SEEK_STATUS_END, nil
				} else {
					//if (DEBUG) {
					//System.out.println("  return END");
					//}
					return index.SEEK_STATUS_END, nil
				}
			} else {
				//if (DEBUG) {
				//System.out.println("  return " + result + " term=" + brToString(term));
				//}
				return result, nil
			}
		} else {
			// Follow this arc
			s.term[targetUpto] = targetLabel
			arc = nextArc
			// Aggregate output as we go:
			//assert arc.output() != null;
			if len(arc.Output()) != 0 {
				output, err = FST_OUTPUTS.Add(output, arc.Output())
				if err != nil {
					return index.SEEK_STATUS_UNDEFINED, err
				}
			}

			//if (DEBUG) {
			//System.out.println("    index: follow label=" + (target.bytes[target.offset + targetUpto]&0xff) + " arc.output=" + arc.output + " arc.nfo=" + arc.nextFinalOutput);
			//}
			targetUpto++

			if arc.IsFinal() {
				//if (DEBUG) System.out.println("    arc is final!");
				frameData, err := FST_OUTPUTS.Add(output, arc.NextFinalOutput())
				if err != nil {
					return index.SEEK_STATUS_UNDEFINED, err
				}
				s.currentFrame, err = s.pushFrameSeekTo(ctx, arc, frameData, targetUpto)
				if err != nil {
					return index.SEEK_STATUS_UNDEFINED, err
				}
				//if (DEBUG) System.out.println("    curFrame.ord=" + currentFrame.ord + " hasTerms=" + currentFrame.hasTerms);
			}
		}
	}

	//validIndexPrefix = targetUpto;
	s.validIndexPrefix = s.currentFrame.prefix

	if err := s.currentFrame.scanToFloorFrame(ctx, target); err != nil {
		return index.SEEK_STATUS_UNDEFINED, err
	}

	if err := s.currentFrame.loadBlock(ctx); err != nil {
		return index.SEEK_STATUS_UNDEFINED, err
	}

	result, err := s.currentFrame.scanToTerm(ctx, target, false)
	if err != nil {
		return index.SEEK_STATUS_UNDEFINED, err
	}

	if result == index.SEEK_STATUS_END {
		copy(s.term, target)
		//s.term.copyBytes(target)
		s.termExists = false
		if _, err := s.Next(ctx); err == nil {
			//if (DEBUG) {
			//System.out.println("  return NOT_FOUND term=" + term.get().utf8ToString() + " " + term);
			//}
			return index.SEEK_STATUS_NOT_FOUND, nil
		} else {
			//if (DEBUG) {
			//System.out.println("  return END");
			//}
			return index.SEEK_STATUS_END, nil
		}
	} else {
		return result, nil
	}
}

func (s *SegmentTermsEnum) SeekExact(ctx context.Context, target []byte) (bool, error) {
	if s.fr.index == nil {
		return false, errors.New("terms index was not loaded")
	}

	frSize, err := s.fr.Size()
	if err != nil {
		return false, err
	}

	getMin, err := s.fr.GetMin()
	if err != nil {
		return false, err
	}
	getMax, err := s.fr.GetMax()
	if err != nil {
		return false, err
	}
	if frSize > 0 && (bytes.Compare(target, getMin)) < 0 || bytes.Compare(target, getMax) > 0 {
		return false, nil
	}

	s.term = ensureSize(s.term, len(target))

	// if (DEBUG) {
	//   System.out.println("\nBTTR.seekExact seg=" + fr.parent.segment + " target=" + fr.fieldInfo.name + ":" + brToString(target) + " current=" + brToString(term) + " (exists?=" + termExists + ") validIndexPrefix=" + validIndexPrefix);
	//   printSeekState(System.out);
	// }

	var arc *fst.Arc[[]byte]
	var targetUpto int
	var output []byte

	s.targetBeforeCurrentLength = s.currentFrame.ord

	if s.currentFrame != s.staticFrame {

		// We are already seek'd; find the common
		// prefix of new seek term vs current term and
		// re-use the corresponding seek state.  For
		// example, if app first seeks to foobar, then
		// seeks to foobaz, we can re-use the seek state
		// for the first 5 bytes.

		// if (DEBUG) {
		//   System.out.println("  re-use current seek state validIndexPrefix=" + validIndexPrefix);
		// }

		arc = s.arcs[0]
		//assert arc.isFinal();
		output = arc.Output()
		targetUpto = 0

		lastFrame := s.stack[0]
		//assert validIndexPrefix <= term.length();

		targetLimit := min(len(target), s.validIndexPrefix)

		cmp := 0

		// TODO: reverse vLong byte order for better FST
		// prefix output sharing

		// First compare up to valid seek frames:
		for targetUpto < targetLimit {
			cmp = int(s.term[targetUpto] - target[targetUpto])
			// if (DEBUG) {
			//    System.out.println("    cycle targetUpto=" + targetUpto + " (vs limit=" + targetLimit + ") cmp=" + cmp + " (targetLabel=" + (char) (target.bytes[target.offset + targetUpto]) + " vs termLabel=" + (char) (term.bytes[targetUpto]) + ")"   + " arc.output=" + arc.output + " output=" + output);
			// }
			if cmp != 0 {
				break
			}
			arc = s.arcs[1+targetUpto]
			//assert arc.label() == (target.bytes[target.offset + targetUpto] & 0xFF): "arc.label=" + (char) arc.label() + " targetLabel=" + (char) (target.bytes[target.offset + targetUpto] & 0xFF);
			if len(arc.Output()) != 0 {
				output, err = FST_OUTPUTS.Add(output, arc.Output())
				if err != nil {
					return false, err
				}
			}
			if arc.IsFinal() {
				lastFrame = s.stack[1+lastFrame.ord]
			}
			targetUpto++
		}

		if cmp == 0 {
			targetUptoMid := targetUpto

			// Second compare the rest of the term, but
			// don't save arc/output/frame; we only do this
			// to find out if the target term is before,
			// equal or after the current term
			targetLimit2 := min(len(target), len(s.term))
			for targetUpto < targetLimit2 {
				cmp = int(s.term[targetUpto] - target[targetUpto])
				// if (DEBUG) {
				//    System.out.println("    cycle2 targetUpto=" + targetUpto + " (vs limit=" + targetLimit + ") cmp=" + cmp + " (targetLabel=" + (char) (target.bytes[target.offset + targetUpto]) + " vs termLabel=" + (char) (term.bytes[targetUpto]) + ")");
				// }
				if cmp != 0 {
					break
				}
				targetUpto++
			}

			if cmp == 0 {
				cmp = len(s.term) - len(target)
			}
			targetUpto = targetUptoMid
		}

		if cmp < 0 {
			// Common case: target term is after current
			// term, ie, app is seeking multiple terms
			// in sorted order
			// if (DEBUG) {
			//   System.out.println("  target is after current (shares prefixLen=" + targetUpto + "); frame.ord=" + lastFrame.ord);
			// }
			s.currentFrame = lastFrame

		} else if cmp > 0 {
			// Uncommon case: target term
			// is before current term; this means we can
			// keep the currentFrame but we must rewind it
			// (so we scan from the start)
			s.targetBeforeCurrentLength = lastFrame.ord
			// if (DEBUG) {
			//   System.out.println("  target is before current (shares prefixLen=" + targetUpto + "); rewind frame ord=" + lastFrame.ord);
			// }
			s.currentFrame = lastFrame
			if err := s.currentFrame.rewind(ctx); err != nil {
				return false, err
			}
		} else {
			// Target is exactly the same as current term
			//assert term.length() == target.length;
			if s.termExists {
				// if (DEBUG) {
				//   System.out.println("  target is same as current; return true");
				// }
				return true, nil
			} else {
				// if (DEBUG) {
				//   System.out.println("  target is same as current but term doesn't exist");
				// }
			}
			//validIndexPrefix = currentFrame.depth;
			//term.length = target.length;
			//return termExists;
		}

	} else {

		s.targetBeforeCurrentLength = -1
		arc, err = s.fr.index.GetFirstArc(s.arcs[0])
		if err != nil {
			return false, err
		}

		// Empty string prefix must have an output (block) in the index!
		//assert arc.isFinal();
		//assert arc.output() != null;

		// if (DEBUG) {
		//   System.out.println("    no seek state; push root frame");
		// }

		output = arc.Output()

		s.currentFrame = s.staticFrame

		//term.length = 0;
		targetUpto = 0
		frameData, err := FST_OUTPUTS.Add(output, arc.NextFinalOutput())
		if err != nil {
			return false, err
		}
		s.currentFrame, err = s.pushFrameSeekTo(ctx, arc, frameData, 0)
		if err != nil {
			return false, err
		}
	}

	// if (DEBUG) {
	//   System.out.println("  start index loop targetUpto=" + targetUpto + " output=" + output + " currentFrame.ord=" + currentFrame.ord + " targetBeforeCurrentLength=" + targetBeforeCurrentLength);
	// }

	// We are done sharing the common prefix with the incoming target and where we are currently seek'd; now continue walking the index:
	for targetUpto < len(target) {

		targetLabel := int(target[targetUpto])

		nextArc, ok, err := s.fr.index.FindTargetArc(ctx, targetLabel, s.fstReader, arc, s.getArc(1+targetUpto))
		if err != nil {
			return false, err
		}

		if !ok {

			// Index is exhausted
			// if (DEBUG) {
			//   System.out.println("    index: index exhausted label=" + ((char) targetLabel) + " " + toHex(targetLabel));
			// }

			s.validIndexPrefix = s.currentFrame.prefix
			//validIndexPrefix = targetUpto;

			if err := s.currentFrame.scanToFloorFrame(ctx, target); err != nil {
				return false, err
			}

			if !s.currentFrame.hasTerms {
				s.termExists = false
				s.term[targetUpto] = byte(targetLabel)
				s.term = s.term[:1+targetUpto]
				// if (DEBUG) {
				//   System.out.println("  FAST NOT_FOUND term=" + brToString(term));
				// }
				return false, nil
			}

			if err := s.currentFrame.loadBlock(ctx); err != nil {
				return false, err
			}

			result, err := s.currentFrame.scanToTerm(ctx, target, true)
			if err != nil {
				return false, err
			}
			if result == index.SEEK_STATUS_FOUND {
				// if (DEBUG) {
				//   System.out.println("  return FOUND term=" + term.utf8ToString() + " " + term);
				// }
				return true, nil
			} else {
				// if (DEBUG) {
				//   System.out.println("  got " + result + "; return NOT_FOUND term=" + brToString(term));
				// }
				return false, nil
			}
		} else {
			// Follow this arc
			arc = nextArc
			s.term[targetUpto] = byte(targetLabel)
			// Aggregate output as we go:
			//assert arc.output() != null;
			if FST_OUTPUTS.IsNoOutput(arc.Output()) {
				output, err = FST_OUTPUTS.Add(output, arc.Output())
				if err != nil {
					return false, err
				}
			}

			// if (DEBUG) {
			//   System.out.println("    index: follow label=" + toHex(target.bytes[target.offset + targetUpto]&0xff) + " arc.output=" + arc.output + " arc.nfo=" + arc.nextFinalOutput);
			// }
			targetUpto++

			if arc.IsFinal() {
				//if (DEBUG) System.out.println("    arc is final!");
				frameData, err := FST_OUTPUTS.Add(output, arc.NextFinalOutput())
				if err != nil {
					return false, err
				}
				s.currentFrame, err = s.pushFrameSeekTo(ctx, arc, frameData, targetUpto)
				if err != nil {
					return false, err
				}
				//if (DEBUG) System.out.println("    curFrame.ord=" + currentFrame.ord + " hasTerms=" + currentFrame.hasTerms);
			}
		}
	}

	//validIndexPrefix = targetUpto;
	s.validIndexPrefix = s.currentFrame.prefix

	if err := s.currentFrame.scanToFloorFrame(ctx, target); err != nil {
		return false, err
	}

	// Target term is entirely contained in the index:
	if !s.currentFrame.hasTerms {
		s.termExists = false
		s.term = s.term[:targetUpto]
		// if (DEBUG) {
		//   System.out.println("  FAST NOT_FOUND term=" + brToString(term));
		// }
		return false, nil
	}

	if err := s.currentFrame.loadBlock(ctx); err != nil {
		return false, err
	}

	result, err := s.currentFrame.scanToTerm(ctx, target, true)
	if err != nil {
		return false, err
	}
	if result == index.SEEK_STATUS_FOUND {
		// if (DEBUG) {
		//   System.out.println("  return FOUND term=" + term.utf8ToString() + " " + term);
		// }
		return true, nil
	} else {
		// if (DEBUG) {
		//   System.out.println("  got result " + result + "; return NOT_FOUND term=" + term.utf8ToString());
		// }

		return false, nil
	}
}

func (s *SegmentTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	return errors.New("unsupported operation")
}

func (s *SegmentTermsEnum) Term() ([]byte, error) {
	return s.term, nil
}

func (s *SegmentTermsEnum) Ord() (int64, error) {
	return 0, errors.New("unsupported operation")
}

func (s *SegmentTermsEnum) DocFreq() (int, error) {
	err := s.currentFrame.decodeMetaData(context.Background())
	if err != nil {
		return 0, err
	}
	return s.currentFrame.state.GetDocFreq(), nil
}

func (s *SegmentTermsEnum) TotalTermFreq() (int64, error) {
	err := s.currentFrame.decodeMetaData(context.Background())
	if err != nil {
		return 0, err
	}
	return int64(s.currentFrame.state.GetTotalTermFreq()), nil
}

func (s *SegmentTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	err := s.currentFrame.decodeMetaData(context.Background())
	if err != nil {
		return nil, err
	}
	return s.fr.parent.postingsReader.Postings(context.Background(), s.fr.fieldInfo, s.currentFrame.state, reuse, flags)
}

func (s *SegmentTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	err := s.currentFrame.decodeMetaData(context.Background())
	if err != nil {
		return nil, err
	}
	return s.fr.parent.postingsReader.Impacts(context.Background(), s.fr.fieldInfo, s.currentFrame.state, flags)
}

func (s *SegmentTermsEnum) initIndexInput() {
	if s.in == nil {
		s.in = s.fr.parent.termsIn.Clone().(store.IndexInput)
	}
}

func (s *SegmentTermsEnum) pushFrameSeekTo(ctx context.Context, arc *fst.Arc[[]byte], frameData []byte, length int) (*SegmentTermsEnumFrame, error) {
	s.scratchReader.Reset(frameData)
	code, err := s.scratchReader.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}

	fpSeek := int64(code >> OUTPUT_FLAGS_NUM_BITS)
	f, err := s.getFrame(1 + s.currentFrame.ord)
	if err != nil {
		return nil, err
	}
	f.hasTerms = (code & OUTPUT_FLAG_HAS_TERMS) != 0
	f.hasTermsOrig = f.hasTerms
	f.isFloor = (code & OUTPUT_FLAG_IS_FLOOR) != 0
	if f.isFloor {
		if err := f.setFloorData(ctx, s.scratchReader, frameData); err != nil {
			return nil, err
		}
	}
	if _, err := s.pushFrame(ctx, arc, fpSeek, length); err != nil {
		return nil, err
	}
	return f, nil
}

// Pushes next'd frame or seek'd frame; we later
// lazy-load the frame only when needed
func (s *SegmentTermsEnum) pushFrame(ctx context.Context, arc *fst.Arc[[]byte],
	fp int64, length int) (*SegmentTermsEnumFrame, error) {

	f, err := s.getFrame(1 + s.currentFrame.ord)
	if err != nil {
		return nil, err
	}
	f.arc = arc
	if f.fpOrig == fp && f.nextEnt != -1 {
		slog.Debug("push reused frame", "ord", f.ord, "fp", f.fp,
			"isFloor", f.isFloor, "hasTerms", f.hasTerms, "pref", string(s.term), "nextEnt", f.nextEnt,
			"targetBeforeCurrentLength", s.targetBeforeCurrentLength, "term.length", len(s.term), "prefix", f.prefix)

		//if (DEBUG) System.out.println("      push reused frame ord=" + f.ord + " fp=" + f.fp + " isFloor?=" + f.isFloor + " hasTerms=" + f.hasTerms + " pref=" + term + " nextEnt=" + f.nextEnt + " targetBeforeCurrentLength=" + targetBeforeCurrentLength + " term.length=" + term.length + " vs prefix=" + f.prefix);
		//if (f.prefix > targetBeforeCurrentLength) {
		if f.ord > s.targetBeforeCurrentLength {
			if err := f.rewind(ctx); err != nil {
				return nil, err
			}
		} else {
			slog.Debug("skip rewind!")
		}
		//assert length == f.prefix;
	} else {
		f.nextEnt = -1
		f.prefix = length
		f.state.SetTermBlockOrd(0)
		f.fpOrig = fp
		f.fp = fp
		f.lastSubFP = -1

		slog.Debug("push reused frame", "ord", f.ord, "fp", f.fp, "hasTerms", f.hasTerms,
			"isFloor", f.isFloor, "pref", string(s.term))
	}

	return f, nil
}

func (s *SegmentTermsEnum) getFrame(ord int) (*SegmentTermsEnumFrame, error) {
	if ord >= len(s.stack) {
		size := ord + 1
		for i := len(s.stack); i < size; i++ {
			frame, err := NewSegmentTermsEnumFrame(s, i)
			if err != nil {
				return nil, err
			}
			s.stack = append(s.stack, frame)
		}
	}
	return s.stack[ord], nil
}

func (s *SegmentTermsEnum) getArc(ord int) *fst.Arc[[]byte] {
	if ord >= len(s.arcs) {
		size := ord + 1
		for i := len(s.arcs); i < size; i++ {
			s.arcs = append(s.arcs, new(fst.Arc[[]byte]))
		}
	}
	return s.arcs[ord]
}
