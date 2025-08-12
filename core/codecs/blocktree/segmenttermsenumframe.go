package blocktree

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/array"
	"github.com/geange/lucene-go/core/util/fst"
	"golang.org/x/exp/slog"
)

type SegmentTermsEnumFrame struct {
	// Our index in stack[]:
	ord int

	hasTerms     bool
	hasTermsOrig bool
	isFloor      bool

	arc *fst.Arc[[]byte]

	//static boolean DEBUG = BlockTreeTermsWriter.DEBUG;

	// File pointer where this block was loaded from
	fp               int64
	fpOrig           int64
	fpEnd            int64
	totalSuffixBytes int

	suffixBytes    []byte
	suffixesReader *store.ByteArrayDataInput

	suffixLengthBytes       []byte
	suffixLengthsReader     *store.ByteArrayDataInput
	statBytes               []byte
	statsSingletonRunLength int
	statsReader             *store.ByteArrayDataInput

	floorData       []byte
	floorDataReader *store.ByteArrayDataInput

	// Length of prefix shared by all terms in this block
	prefix int

	// Number of entries (term or sub-block) in this block
	entCount int

	// Which term we will next read, or -1 if the block
	// isn't loaded yet
	nextEnt int

	// True if this block is either not a floor block,
	// or, it's the last sub-block of a floor block
	isLastInFloor bool

	// True if all entries are terms
	isLeafBlock bool

	lastSubFP int64

	nextFloorLabel       int
	numFollowFloorBlocks int

	// Next term to decode metaData; we decode metaData
	// lazily so that scanning to find the matching term is
	// fast and only if you find a match and app wants the
	// stats or docs/positions enums, will we decode the
	// metaData
	metaDataUpto int

	state types.BlockTermState

	// metadata buffer
	bytes       []byte
	bytesReader *store.ByteArrayDataInput

	ste     *SegmentTermsEnum
	version int

	startBytePos int
	suffix       int
	subCode      int64
	//compressionAlg CompressionAlgorithm
}

func NewSegmentTermsEnumFrame(ste *SegmentTermsEnum, ord int) (*SegmentTermsEnumFrame, error) {
	this := &SegmentTermsEnumFrame{
		suffixBytes:     make([]byte, 128),
		suffixesReader:  store.NewByteArrayDataInput(nil),
		statBytes:       make([]byte, 64),
		statsReader:     store.NewByteArrayDataInput(nil),
		floorData:       make([]byte, 32),
		floorDataReader: store.NewByteArrayDataInput(nil),
		bytes:           make([]byte, 32),
		bytesReader:     store.NewByteArrayDataInput(nil),
	}
	this.ste = ste
	this.ord = ord
	state, err := ste.fr.parent.postingsReader.NewTermState()
	if err != nil {
		return nil, err
	}
	this.state = state
	this.state.SetTotalTermFreq(-1)
	this.version = ste.fr.parent.version
	if this.version >= VERSION_COMPRESSED_SUFFIXES {
		this.suffixLengthBytes = make([]byte, 32)
		this.suffixLengthsReader = store.NewByteArrayDataInput(nil)
	} else {
		this.suffixLengthBytes = nil
		this.suffixLengthsReader = this.suffixesReader
	}
	return this, nil
}

func (s *SegmentTermsEnumFrame) setFloorData(ctx context.Context, in *store.ByteArrayDataInput, source []byte) error {
	numBytes := len(source) - (in.GetPosition())
	if numBytes > len(s.floorData) {
		s.floorData = make([]byte, numBytes)
	}
	copy(s.floorData, source)
	s.floorDataReader.Reset(s.floorData)

	numBlocks, err := s.floorDataReader.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	s.numFollowFloorBlocks = int(numBlocks)

	label, err := s.floorDataReader.ReadByte()
	if err != nil {
		return err
	}
	s.nextFloorLabel = int(label)
	return nil
}

func (s *SegmentTermsEnumFrame) getTermBlockOrd() int {
	if s.isLeafBlock {
		return s.nextEnt
	}
	return s.state.GetTermBlockOrd()
}

func (s *SegmentTermsEnumFrame) loadNextFloorBlock(ctx context.Context) error {
	s.fp = s.fpEnd
	s.nextEnt = -1
	return s.loadBlock(ctx)
}

/*
Does initial decode of next block of terms; this
doesn't actually decode the docFreq, totalTermFreq,
postings details (frq/prx offset, etc.) metadata;
it just loads them as byte[] blobs which are then
decoded on-demand if the metadata is ever requested
for any term in this block.  This enables terms-only
intensive consumes (eg certain MTQs, respelling) to
not pay the price of decoding metadata they won't
use.
*/
func (s *SegmentTermsEnumFrame) loadBlock(ctx context.Context) error {
	// Clone the IndexInput lazily, so that consumers
	// that just pull a TermsEnum to
	// seekExact(TermState) don't pay this cost:
	s.ste.initIndexInput()

	if s.nextEnt != -1 {
		// Already loaded
		return nil
	}
	//System.out.println("blc=" + blockLoadCount);

	if _, err := s.ste.in.Seek(s.fp, io.SeekStart); err != nil {
		return err
	}
	code, err := s.ste.in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	s.entCount = int(code >> 1)
	s.isLastInFloor = (code & 1) != 0

	// TODO: if suffixes were stored in random-access
	// array structure, then we could do binary search
	// instead of linear scan to find target term; eg
	// we could have simple array of offsets

	startSuffixFP := s.ste.in.GetFilePointer()
	// term suffixes:
	if s.version >= VERSION_COMPRESSED_SUFFIXES {
		codeL, err := s.ste.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		s.isLeafBlock = (codeL & 0x04) != 0
		numSuffixBytes := (int)(codeL >> 3)
		if len(s.suffixBytes) < numSuffixBytes {
			s.suffixBytes = make([]byte, array.Oversize(numSuffixBytes, 1))
		}

		compressionAlg, err := ByCode(int(codeL & 0x03))
		if err != nil {
			return err
		}

		if err := compressionAlg.Read(ctx, s.ste.in, s.suffixBytes[:numSuffixBytes]); err != nil {
			return err
		}
		s.suffixesReader.Reset(s.suffixBytes[:numSuffixBytes])

		numSuffixLengthBytes, err := s.ste.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		allEqual := (numSuffixLengthBytes & 0x01) != 0
		numSuffixLengthBytes >>= 1
		if len(s.suffixLengthBytes) < int(numSuffixLengthBytes) {
			s.suffixLengthBytes = make([]byte, array.Oversize(int(numSuffixLengthBytes), 1))
		}
		if allEqual {
			b, err := s.ste.in.ReadByte()
			if err != nil {
				return err
			}
			array.Fill(s.suffixLengthBytes[:numSuffixLengthBytes], b)
		} else {
			if _, err := s.ste.in.Read(s.suffixLengthBytes[:numSuffixLengthBytes]); err != nil {
				return err
			}
		}
		s.suffixLengthsReader.Reset(s.suffixLengthBytes[:numSuffixLengthBytes])
	} else {
		code, err = s.ste.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		s.isLeafBlock = (code & 1) != 0
		numBytes := int(code >> 1)
		if len(s.suffixBytes) < numBytes {
			s.suffixBytes = make([]byte, array.Oversize(numBytes, 1))
		}
		if _, err := s.ste.in.Read(s.suffixBytes[:numBytes]); err != nil {
			return err
		}
		s.suffixesReader.Reset(s.suffixBytes[:numBytes])
	}
	s.totalSuffixBytes = int(s.ste.in.GetFilePointer() - startSuffixFP)

	// stats
	numBytesInt64, err := s.ste.in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	numBytes := int(numBytesInt64)
	if len(s.statBytes) < int(numBytes) {
		s.statBytes = make([]byte, array.Oversize(numBytes, 1))
	}
	s.ste.in.Read(s.statBytes[:numBytes])
	s.statsReader.Reset(s.statBytes[:numBytes])
	s.statsSingletonRunLength = 0
	s.metaDataUpto = 0

	s.state.SetTermBlockOrd(0)
	s.nextEnt = 0
	s.lastSubFP = -1

	// TODO: we could skip this if !hasTerms; but
	// that's rare so won't help much
	// metadata
	numBytesInt64, err = s.ste.in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	numBytes = int(numBytesInt64)
	if len(s.bytes) < numBytes {
		s.bytes = make([]byte, array.Oversize(numBytes, 1))
	}
	s.ste.in.Read(s.bytes[:numBytes])
	s.bytesReader.Reset(s.bytes[:numBytes])

	// Sub-blocks of a single floor block are always
	// written one after another -- tail recurse:
	s.fpEnd = s.ste.in.GetFilePointer()
	return nil
}

func (s *SegmentTermsEnumFrame) rewind(ctx context.Context) error {
	s.fp = s.fpOrig
	s.nextEnt = -1
	s.hasTerms = s.hasTermsOrig
	if s.isFloor {
		s.floorDataReader.Rewind()
		numFollowFloorBlocks, err := s.floorDataReader.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		s.numFollowFloorBlocks = int(numFollowFloorBlocks)

		nextFloorLabel, err := s.floorDataReader.ReadByte()
		if err != nil {
			return err
		}
		s.nextFloorLabel = int(nextFloorLabel)
	}
	return nil
}

func (s *SegmentTermsEnumFrame) next(ctx context.Context) (bool, error) {
	if s.isLeafBlock {
		if err := s.nextLeaf(ctx); err != nil {
			return false, err
		}
		return false, nil
	} else {
		return s.nextNonLeaf(ctx)
	}
}

func (s *SegmentTermsEnumFrame) nextLeaf(ctx context.Context) error {
	s.nextEnt++
	suffix, err := s.suffixLengthsReader.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	s.suffix = int(suffix)

	s.startBytePos = s.suffixesReader.GetPosition()

	s.ste.term = array.Grow(s.ste.term, s.prefix+s.suffix)
	s.ste.term = s.ste.term[:s.prefix+s.suffix]
	if _, err = s.suffixesReader.Read(s.ste.term[s.prefix:s.suffix]); err != nil {
		return err
	}

	s.ste.termExists = true
	return nil
}

func (s *SegmentTermsEnumFrame) nextNonLeaf(ctx context.Context) (bool, error) {
	for {
		if s.nextEnt == s.entCount {
			//assert arc == null || (isFloor && isLastInFloor == false): "isFloor=" + isFloor + " isLastInFloor=" + isLastInFloor;
			if err := s.loadNextFloorBlock(ctx); err != nil {
				return false, err
			}
			if s.isLeafBlock {
				if err := s.nextLeaf(ctx); err != nil {
					return false, err
				}
				return false, nil
			} else {
				continue
			}
		}

		//assert nextEnt != -1 && nextEnt < entCount: "nextEnt=" + nextEnt + " entCount=" + entCount + " fp=" + fp;
		s.nextEnt++
		code, err := s.suffixLengthsReader.ReadUvarint(ctx)
		if err != nil {
			return false, err
		}
		s.suffix = int(code >> 1)
		s.startBytePos = s.suffixesReader.GetPosition()
		termSize := s.prefix + s.suffix
		s.ste.term = array.Grow(s.ste.term, termSize)
		s.ste.term = s.ste.term[:termSize]
		if _, err := s.suffixesReader.Read(s.ste.term); err != nil {
			return false, err
		}
		if (code & 1) == 0 {
			// A normal term
			s.ste.termExists = true
			s.subCode = 0
			s.state.AddTermBlockOrd(1)
			return false, nil
		} else {
			// A sub-block; make sub-FP absolute:
			s.ste.termExists = false
			subCode, err := s.suffixLengthsReader.ReadUvarint(ctx)
			if err != nil {
				return false, err
			}
			s.subCode = int64(subCode)
			s.lastSubFP = s.fp - s.subCode
			//if (DEBUG) {
			//System.out.println("    lastSubFP=" + lastSubFP);
			//}
			return true, nil
		}
	}
}

func (s *SegmentTermsEnumFrame) scanToFloorFrame(ctx context.Context, target []byte) error {
	if !s.isFloor || len(target) <= s.prefix {
		slog.Debug("scanToFloorFrame skip", "isFloor", s.isFloor, "target.length", len(target), "prefix", s.prefix)
		return nil
	}

	targetLabel := target[s.prefix]

	slog.Debug("scanToFloorFrame", "fpOrig", s.fpOrig,
		"targetLabel", fmt.Sprintf("%x", targetLabel),
		"nextFloorLabel", fmt.Sprintf("%x", s.nextFloorLabel),
		"numFollowFloorBlocks", s.numFollowFloorBlocks)

	if int(targetLabel) < s.nextFloorLabel {
		slog.Debug("already on correct block")
		return nil
	}

	newFP := s.fpOrig
	for {
		code, err := s.floorDataReader.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		newFP = s.fpOrig + int64(code>>1)
		s.hasTerms = (code & 1) != 0

		slog.Debug("loop print",
			"label", fmt.Sprintf("%X", s.nextFloorLabel), "fp", newFP,
			"hasTerms", s.hasTerms, "numFollowFloor", s.numFollowFloorBlocks)

		s.isLastInFloor = s.numFollowFloorBlocks == 1
		s.numFollowFloorBlocks--

		if s.isLastInFloor {
			s.nextFloorLabel = 256

			slog.Debug("stop!", "nextFloorLabel", fmt.Sprintf("%X", s.nextFloorLabel))
			break
		} else {
			nextFloorLabel, err := s.floorDataReader.ReadByte()
			if err != nil {
				return err
			}
			s.nextFloorLabel = int(nextFloorLabel)
			if targetLabel < nextFloorLabel {
				slog.Debug("stop!", "nextFloorLabel", fmt.Sprintf("%X", s.nextFloorLabel))
				break
			}
		}
	}

	if newFP != s.fp {
		slog.Debug("force switch to", "fp", newFP, "oldFP", s.fp)
		s.nextEnt = -1
		s.fp = newFP
	} else {
		slog.Debug("stay on same", "fp", newFP)
	}
	return nil
}

func (s *SegmentTermsEnumFrame) decodeMetaData(ctx context.Context) error {
	// lazily catch up on metadata decode:
	limit := s.getTermBlockOrd()
	absolute := s.metaDataUpto == 0

	// TODO: better API would be "jump straight to term=N"???
	for s.metaDataUpto < limit {
		// TODO: we could make "tiers" of metadata, ie,
		// decode docFreq/totalTF but don't decode postings
		// metadata; this way caller could get
		// docFreq/totalTF w/o paying decode cost for
		// postings

		// TODO: if docFreq were bulk decoded we could
		// just skipN here:

		if s.version >= VERSION_COMPRESSED_SUFFIXES {
			if s.statsSingletonRunLength > 0 {
				s.state.SetDocFreq(1)
				s.state.SetTotalTermFreq(1)
				s.statsSingletonRunLength--
			} else {
				token, err := s.statsReader.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				if (token & 1) == 1 {
					s.state.SetDocFreq(1)
					s.state.SetTotalTermFreq(1)
					s.statsSingletonRunLength = int(token >> 1)
				} else {
					s.state.SetDocFreq(int(token >> 1))
					if s.ste.fr.fieldInfo.GetIndexOptions() == document.INDEX_OPTIONS_DOCS {
						s.state.SetTotalTermFreq(s.state.GetDocFreq())
					} else {
						freq, err := s.statsReader.ReadUvarint(ctx)
						if err != nil {
							return err
						}
						s.state.SetTotalTermFreq(s.state.GetDocFreq() + int(freq))
					}
				}
			}
		} else {
			docFreq, err := s.statsReader.ReadUvarint(ctx)
			if err != nil {
				return err
			}

			s.state.SetDocFreq(int(docFreq))
			if s.ste.fr.fieldInfo.GetIndexOptions() == document.INDEX_OPTIONS_DOCS {
				s.state.SetTotalTermFreq(s.state.GetDocFreq()) // all postings have freq=1
			} else {
				n, err := s.statsReader.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				s.state.SetTotalTermFreq(s.state.GetDocFreq() + int(n))
			}
		}

		// metadata
		err := s.ste.fr.parent.postingsReader.DecodeTerm(ctx, s.bytesReader, s.ste.fr.fieldInfo, s.state, absolute)
		if err != nil {
			return err
		}

		s.metaDataUpto++
		absolute = false
	}
	s.state.SetTermBlockOrd(s.metaDataUpto)
	return nil
}

func (s *SegmentTermsEnumFrame) prefixMatches(target []byte) bool {
	return bytes.HasSuffix(target, s.ste.term[:s.prefix])
}

// Scans to sub-block that has this target fp; only
// called by next(); NOTE: does not set
// startBytePos/suffix as a side effect
func (s *SegmentTermsEnumFrame) scanToSubBlock(ctx context.Context, subFP int64) error {
	slog.Debug("scanToSubBlock", "fp", s.fp, "subFP", subFP, "entCount", s.entCount, "lastSubFP", s.lastSubFP)
	//assert nextEnt == 0;
	if s.lastSubFP == subFP {
		//if (DEBUG) System.out.println("    already positioned");
		return nil
	}
	//assert subFP < fp : "fp=" + fp + " subFP=" + subFP;
	targetSubCode := s.fp - subFP
	//if (DEBUG) System.out.println("    targetSubCode=" + targetSubCode);
	for {
		//assert nextEnt < entCount;
		s.nextEnt++
		code, err := s.suffixLengthsReader.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		if err := s.suffixesReader.SkipBytes(ctx, int(code>>1)); err != nil {
			return err
		}
		if (code & 1) != 0 {
			subCode, err := s.suffixLengthsReader.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			if targetSubCode == int64(subCode) {
				//if (DEBUG) System.out.println("        match!");
				s.lastSubFP = subFP
				return nil
			}
		} else {
			s.state.AddTermBlockOrd(1)
		}
	}
}

func (s *SegmentTermsEnumFrame) scanToTerm(ctx context.Context, target []byte, exactOnly bool) (coreIndex.SeekStatus, error) {
	if s.isLeafBlock {
		return s.scanToTermLeaf(ctx, target, exactOnly)
	}
	return s.scanToTermNonLeaf(ctx, target, exactOnly)
}

func (s *SegmentTermsEnumFrame) scanToTermLeaf(ctx context.Context, target []byte, exactOnly bool) (coreIndex.SeekStatus, error) {
	// if (DEBUG) System.out.println("    scanToTermLeaf: block fp=" + fp + " prefix=" + prefix + " nextEnt=" + nextEnt + " (of " + entCount + ") target=" + brToString(target) + " term=" + brToString(term));

	//assert nextEnt != -1;

	s.ste.termExists = true
	s.subCode = 0

	if s.nextEnt == s.entCount {
		if exactOnly {
			s.fillTerm()
		}
		return coreIndex.SEEK_STATUS_END, nil
	}

	//assert prefixMatches(target);

	// TODO: binary search when all terms have the same length, which is common for ID fields,
	// which are also the most sensitive to lookup performance?
	// Loop over each entry (term or sub-block) in this block:
	for {
		s.nextEnt++

		suffix, err := s.suffixLengthsReader.ReadUvarint(ctx)
		if err != nil {
			return coreIndex.SEEK_STATUS_UNDEFINED, err
		}

		s.suffix = int(suffix)

		// if (DEBUG) {
		//   BytesRef suffixBytesRef = new BytesRef();
		//   suffixBytesRef.bytes = suffixBytes;
		//   suffixBytesRef.offset = suffixesReader.getPosition();
		//   suffixBytesRef.length = suffix;
		//   System.out.println("      cycle: term " + (nextEnt-1) + " (of " + entCount + ") suffix=" + brToString(suffixBytesRef));
		// }

		s.startBytePos = s.suffixesReader.GetPosition()
		if err := s.suffixesReader.SkipBytes(ctx, s.suffix); err != nil {
			return coreIndex.SEEK_STATUS_UNDEFINED, err
		}

		// Loop over bytes in the suffix, comparing to the target
		cmp := bytes.Compare(s.suffixBytes[s.startBytePos:s.startBytePos+s.suffix], target)

		if cmp < 0 {
			// Current entry is still before the target;
			// keep scanning
		} else if cmp > 0 {
			// Done!  Current entry is after target --
			// return NOT_FOUND:
			s.fillTerm()

			//if (DEBUG) System.out.println("        not found");
			return coreIndex.SEEK_STATUS_NOT_FOUND, nil
		} else {
			// Exact match!

			// This cannot be a sub-block because we
			// would have followed the index to this
			// sub-block from the start:

			//assert ste.termExists;
			s.fillTerm()
			//if (DEBUG) System.out.println("        found!");
			return coreIndex.SEEK_STATUS_FOUND, nil
		}

		if s.nextEnt < s.entCount {
			break
		}
	}

	// It is possible (and OK) that terms index pointed us
	// at this block, but, we scanned the entire block and
	// did not find the term to position to.  This happens
	// when the target is after the last term in the block
	// (but, before the next term in the index).  EG
	// target could be foozzz, and terms index pointed us
	// to the foo* block, but the last term in this block
	// was fooz (and, eg, first term in the next block will
	// bee fop).
	//if (DEBUG) System.out.println("      block end");
	if exactOnly {
		s.fillTerm()
	}

	// TODO: not consistent that in the
	// not-exact case we don't next() into the next
	// frame here
	return coreIndex.SEEK_STATUS_END, nil
}

func (s *SegmentTermsEnumFrame) scanToTermNonLeaf(ctx context.Context, target []byte, exactOnly bool) (coreIndex.SeekStatus, error) {
	//if (DEBUG) System.out.println("    scanToTermNonLeaf: block fp=" + fp + " prefix=" + prefix + " nextEnt=" + nextEnt + " (of " + entCount + ") target=" + brToString(target) + " term=" + brToString(target));

	//assert nextEnt != -1;

	if s.nextEnt == s.entCount {
		if exactOnly {
			s.fillTerm()
			s.ste.termExists = s.subCode == 0
		}
		return coreIndex.SEEK_STATUS_END, nil
	}

	//assert prefixMatches(target);

	// Loop over each entry (term or sub-block) in this block:
	for s.nextEnt < s.entCount {

		s.nextEnt++

		code, err := s.suffixLengthsReader.ReadUvarint(ctx)
		if err != nil {
			return coreIndex.SEEK_STATUS_UNDEFINED, err
		}
		s.suffix = int(code >> 1)

		//if (DEBUG) {
		//  BytesRef suffixBytesRef = new BytesRef();
		//  suffixBytesRef.bytes = suffixBytes;
		//  suffixBytesRef.offset = suffixesReader.getPosition();
		//  suffixBytesRef.length = suffix;
		//  System.out.println("      cycle: " + ((code&1)==1 ? "sub-block" : "term") + " " + (nextEnt-1) + " (of " + entCount + ") suffix=" + brToString(suffixBytesRef));
		//}

		termLen := s.prefix + s.suffix
		s.startBytePos = s.suffixesReader.GetPosition()
		if err := s.suffixesReader.SkipBytes(ctx, s.suffix); err != nil {
			return coreIndex.SEEK_STATUS_UNDEFINED, err
		}
		s.ste.termExists = (code & 1) == 0
		if s.ste.termExists {
			s.state.AddTermBlockOrd(1)
			s.subCode = 0
		} else {
			subCode, err := s.suffixLengthsReader.ReadUvarint(ctx)
			if err != nil {
				return coreIndex.SEEK_STATUS_UNDEFINED, err
			}
			s.subCode = int64(subCode)
			s.lastSubFP = s.fp - s.subCode
		}

		cmp := bytes.Compare(s.suffixBytes[s.startBytePos:s.startBytePos+s.suffix], target)

		if cmp < 0 {
			// Current entry is still before the target;
			// keep scanning
		} else if cmp > 0 {
			// Done!  Current entry is after target --
			// return NOT_FOUND:
			s.fillTerm()

			//if (DEBUG) System.out.println("        maybe done exactOnly=" + exactOnly + " ste.termExists=" + ste.termExists);

			if !exactOnly && !s.ste.termExists {
				//System.out.println("  now pushFrame");
				// TODO this
				// We are on a sub-block, and caller wants
				// us to position to the next term after
				// the target, so we must recurse into the
				// sub-frame(s):
				currentFrame, err := s.ste.pushFrame(ctx, nil, s.ste.currentFrame.lastSubFP, termLen)
				if err != nil {
					return coreIndex.SEEK_STATUS_UNDEFINED, err
				}
				s.ste.currentFrame = currentFrame
				if err := s.ste.currentFrame.loadBlock(ctx); err != nil {
					return coreIndex.SEEK_STATUS_UNDEFINED, err
				}
				for {
					ok, err := s.ste.currentFrame.next(ctx)
					if err != nil {
						return coreIndex.SEEK_STATUS_UNDEFINED, err
					}
					if !ok {
						break
					}
					frame, err := s.ste.pushFrame(ctx, nil, s.ste.currentFrame.lastSubFP, len(s.ste.term))
					if err != nil {
						return coreIndex.SEEK_STATUS_UNDEFINED, err
					}
					s.ste.currentFrame = frame
					if err := s.ste.currentFrame.loadBlock(ctx); err != nil {
						return coreIndex.SEEK_STATUS_UNDEFINED, err
					}
				}
			}

			//if (DEBUG) System.out.println("        not found");
			return coreIndex.SEEK_STATUS_NOT_FOUND, nil
		} else {
			// Exact match!

			// This cannot be a sub-block because we
			// would have followed the index to this
			// sub-block from the start:

			//assert ste.termExists;
			s.fillTerm()
			//if (DEBUG) System.out.println("        found!");
			return coreIndex.SEEK_STATUS_FOUND, nil
		}
	}

	// It is possible (and OK) that terms index pointed us
	// at this block, but, we scanned the entire block and
	// did not find the term to position to.  This happens
	// when the target is after the last term in the block
	// (but, before the next term in the index).  EG
	// target could be foozzz, and terms index pointed us
	// to the foo* block, but the last term in this block
	// was fooz (and, eg, first term in the next block will
	// bee fop).
	//if (DEBUG) System.out.println("      block end");
	if exactOnly {
		s.fillTerm()
	}

	// TODO: not consistent that in the
	// not-exact case we don't next() into the next
	// frame here
	return coreIndex.SEEK_STATUS_END, nil
}

func (s *SegmentTermsEnumFrame) fillTerm() {
	termLength := s.prefix + s.suffix
	s.ste.term = array.Grow(s.ste.term, termLength)
	s.ste.term = s.ste.term[:termLength]
	copy(s.ste.term[s.prefix:], s.suffixBytes[s.startBytePos:s.startBytePos+s.suffix])
}
