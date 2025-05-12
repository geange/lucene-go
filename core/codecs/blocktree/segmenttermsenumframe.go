package blocktree

import (
	"context"
	"github.com/geange/lucene-go/core/codecs/types"
	coreIndex "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
)

type SegmentTermsEnumFrame struct {
	// Our index in stack[]:
	ord int

	hasTerms     bool
	hasTermsOrig bool
	isFloor      bool

	arc *fst.Arc

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

func (s *SegmentTermsEnumFrame) setFloorData(in *store.ByteArrayDataInput, source []byte) error {
	panic("implement me")
}

func (s *SegmentTermsEnumFrame) getTermBlockOrd() int {
	if s.isLeafBlock {
		return s.nextEnt
	}
	return s.state.GetTermBlockOrd()
}

func (s *SegmentTermsEnumFrame) loadNextFloorBlock() error {
	s.fp = s.fpEnd
	s.nextEnt = -1
	return s.loadBlock()
}

func (s *SegmentTermsEnumFrame) loadBlock() error {
	panic("implement me")
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
	panic("implement me")
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
