package blocktree

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/codecs/types"
	coreIndex "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/array"
	"github.com/geange/lucene-go/core/util/fst"
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
	floorDataReader store.ByteArrayDataInput

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
	bytesReader store.ByteArrayDataInput

	ste     *SegmentTermsEnum
	version int

	startBytePos int
	suffix       int
	subCode      int64
	//compressionAlg CompressionAlgorithm
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
	bs := make([]byte, s.prefix+s.suffix)

	if _, err = s.suffixesReader.Read(bs[s.prefix:s.suffix]); err != nil {
		return err
	}
	s.ste.term.Reset()
	s.ste.term.Write(bs)

	s.ste.termExists = true
	return nil
}

func (s *SegmentTermsEnumFrame) nextNonLeaf(ctx context.Context) (bool, error) {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) scanToFloorFrame(target []byte) error {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) decodeMetaData() error {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) prefixMatches(target []byte) error {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) scanToSubBlock(subFP int64) error {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) scanToTerm(target []byte, exactOnly bool) (coreIndex.SeekStatus, error) {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) scanToTermLeaf(target []byte, exactOnly bool) (coreIndex.SeekStatus, error) {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) scanToTermNonLeaf(target []byte, exactOnly bool) (coreIndex.SeekStatus, error) {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) fillTerm() {

}
