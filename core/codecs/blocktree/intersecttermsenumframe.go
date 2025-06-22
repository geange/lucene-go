package blocktree

import (
	"context"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
)

type IntersectTermsEnumFrame struct {
	ord       int
	fp        int64
	fpOrig    int64
	fpEnd     int64
	lastSubFP int64

	// private static boolean DEBUG = IntersectTermsEnum.DEBUG;

	state                   int // State in automaton
	lastState               int // State just before the last label
	metaDataUpto            int
	suffixBytes             []byte
	suffixesReader          *store.ByteArrayDataInput
	suffixLengthBytes       []byte
	suffixLengthsReader     *store.ByteArrayDataInput
	statBytes               []byte
	statsSingletonRunLength int
	statsReader             *store.ByteArrayDataInput
	floorData               []byte
	floorDataReader         *store.ByteArrayDataInput

	prefix   int // Length of prefix shared by all terms in this block
	entCount int // Number of entries (term or sub-block) in this block
	nextEnt  int // Which term we will next read

	// True if this block is either not a floor block,
	// or, it's the last sub-block of a floor block
	isLastInFloor bool

	isLeafBlock          bool // True if all entries are terms
	numFollowFloorBlocks int
	nextFloorLabel       int
	transitionIndex      int
	transitionCount      int
	arc                  *fst.Arc
	termState            types.BlockTermState
	bytes                []byte // metadata buffer
	bytesReader          *store.ByteArrayDataInput
	outputPrefix         []byte // Cumulative output so far
	startBytePos         int
	suffix               int
	ite                  *IntersectTermsEnum
	version              int
}

func (i *IntersectTermsEnumFrame) GetTermBlockOrd() int {
	if i.isLeafBlock {
		return i.nextEnt
	}
	return i.termState.GetTermBlockOrd()
}

func (i *IntersectTermsEnumFrame) DecodeMetaData() error {
	panic("")
}

func (i *IntersectTermsEnumFrame) load(ctx context.Context, frameIndexData []byte) error {
	panic("")
}
