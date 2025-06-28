package blocktree

import (
	"context"
	"io"
	"regexp"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/array"
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
	regexp               *regexp.Regexp
}

func (i *IntersectTermsEnumFrame) load(ctx context.Context, frameIndexData []byte) error {
	if len(frameIndexData) > 0 {
		i.floorDataReader.Reset(frameIndexData)
		// Skip first long -- has redundant fp, hasTerms
		// flag, isFloor flag
		code, err := i.floorDataReader.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		if (code & OUTPUT_FLAG_IS_FLOOR) != 0 {
			// Floor frame
			numFollowFloorBlocks, err := i.floorDataReader.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			i.numFollowFloorBlocks = int(numFollowFloorBlocks)

			_, err = i.floorDataReader.ReadByte()
			if err != nil {
				return err
			}
			//i.nextFloorLabel = int(nextFloorLabel)
		}
	}

	_, err := i.ite.in.Seek(i.fp, io.SeekStart)
	if err != nil {
		return err
	}
	code, err := i.ite.in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	i.entCount = int(code >> 1)
	i.isLastInFloor = (code & 1) != 0

	if i.version >= VERSION_COMPRESSED_SUFFIXES {
		codeL, err := i.ite.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		i.isLeafBlock = (codeL & 0x04) != 0
		numSuffixBytes := int(codeL >> 3)
		if len(i.suffixBytes) < numSuffixBytes {
			i.suffixBytes = array.Grow(i.suffixBytes, numSuffixBytes)
		}
		compressionAlg, err := ByCode(int(codeL & 0x03))
		if err != nil {
			return err
		}

		err = compressionAlg.Read(ctx, i.ite.in, i.suffixBytes[:numSuffixBytes])
		if err != nil {
			return err
		}
		i.suffixesReader.Reset(i.suffixBytes[:numSuffixBytes])

		numSuffixLengthBytes, err := i.ite.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		allEqual := (numSuffixLengthBytes & 0x01) != 0
		numSuffixLengthBytes = numSuffixLengthBytes >> 1
		if len(i.suffixLengthBytes) < int(numSuffixLengthBytes) {
			i.suffixLengthBytes = array.Grow(i.suffixLengthBytes, int(numSuffixLengthBytes))
		}
		if allEqual {
			b, err := i.ite.in.ReadByte()
			if err != nil {
				return err
			}
			for j := 0; j < int(numSuffixLengthBytes); j++ {
				i.suffixLengthBytes[j] = b
			}
		} else {
			if _, err := i.ite.in.Read(i.suffixLengthBytes[:int(numSuffixLengthBytes)]); err != nil {
				return err
			}
		}
		i.suffixLengthsReader.Reset(i.suffixLengthBytes[:int(numSuffixLengthBytes)])
	} else {
		code, err = i.ite.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		i.isLeafBlock = (code & 1) != 0
		numBytes := int(code >> 1)
		if len(i.suffixBytes) < numBytes {
			i.suffixBytes = array.Grow(i.suffixBytes, numBytes)
		}
		if _, err := i.ite.in.Read(i.suffixBytes[:numBytes]); err != nil {
			return err
		}
		i.suffixesReader.Reset(i.suffixBytes[:numBytes])
	}

	// stats
	numBytes, err := i.ite.in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	if len(i.statBytes) < int(numBytes) {
		i.statBytes = array.Grow(i.statBytes, int(numBytes))
	}
	if _, err = i.ite.in.Read(i.statBytes[:numBytes]); err != nil {
		return err
	}
	i.statsReader.Reset(i.statBytes[:numBytes])
	i.statsSingletonRunLength = 0
	i.metaDataUpto = 0

	i.termState.SetTermBlockOrd(0)
	i.nextEnt = 0

	// metadata
	numBytes, err = i.ite.in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	if len(i.bytes) < int(numBytes) {
		i.bytes = make([]byte, numBytes)
	}
	if _, err = i.ite.in.Read(i.bytes); err != nil {
		return err
	}
	i.bytesReader.Reset(i.bytes)

	if !i.isLastInFloor {
		// Sub-blocks of a single floor block are always
		// written one after another -- tail recurse:
		i.fpEnd = i.ite.in.GetFilePointer()
	}
	return nil
}

func (i *IntersectTermsEnumFrame) NextLeaf(ctx context.Context) error {
	i.nextEnt++
	suffix, err := i.suffixLengthsReader.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	i.suffix = int(suffix)
	i.startBytePos = i.suffixesReader.GetPosition()
	return i.suffixesReader.SkipBytes(ctx, i.suffix)
}

func (i *IntersectTermsEnumFrame) nextNonLeaf(ctx context.Context) (bool, error) {
	i.nextEnt++
	code, err := i.suffixLengthsReader.ReadUvarint(ctx)
	if err != nil {
		return false, err
	}
	i.suffix = int(code >> 1)
	i.startBytePos = i.suffixesReader.GetPosition()
	err = i.suffixesReader.SkipBytes(ctx, i.suffix)
	if err != nil {
		return false, err
	}
	if (code & 1) == 0 {
		// A normal term
		i.termState.SetTermBlockOrd(i.termState.GetTermBlockOrd() + 1)
		return false, nil
	} else {
		// A sub-block; make sub-FP absolute:
		n, err := i.suffixLengthsReader.ReadUvarint(ctx)
		if err != nil {
			return false, err
		}
		i.lastSubFP = i.fp - int64(n)
		return true, nil
	}
}

func (i *IntersectTermsEnumFrame) GetTermBlockOrd() int {
	if i.isLeafBlock {
		return i.nextEnt
	}
	return i.termState.GetTermBlockOrd()
}

func (i *IntersectTermsEnumFrame) DecodeMetaData(ctx context.Context) error {
	// lazily catch up on metadata decode:
	limit := i.GetTermBlockOrd()
	absolute := i.metaDataUpto == 0
	for i.metaDataUpto < limit {
		// TODO: we could make "tiers" of metadata, ie,
		// decode docFreq/totalTF but don't decode postings
		// metadata; this way caller could get
		// docFreq/totalTF w/o paying decode cost for
		// postings

		// TODO: if docFreq were bulk decoded we could
		// just skipN here:

		// stats
		if i.version >= VERSION_COMPRESSED_SUFFIXES {
			if i.statsSingletonRunLength > 0 {
				i.termState.SetDocFreq(1)
				i.termState.SetTotalTermFreq(1)
				i.statsSingletonRunLength--
			} else {
				token, err := i.statsReader.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				if i.version >= VERSION_COMPRESSED_SUFFIXES && (token&1) == 1 {
					i.termState.SetDocFreq(1)
					i.termState.SetTotalTermFreq(1)
					i.statsSingletonRunLength = int(token >> 1)
				} else {
					i.termState.SetDocFreq(int(token >> 1))
					if i.ite.fr.fieldInfo.GetIndexOptions() == document.INDEX_OPTIONS_DOCS {
						i.termState.SetTotalTermFreq(i.termState.GetDocFreq())
					} else {
						n, err := i.statsReader.ReadUvarint(ctx)
						if err != nil {
							return err
						}
						i.termState.SetTotalTermFreq(i.termState.GetDocFreq() + int(n))
					}
				}
			}
		} else {
			docFreq, err := i.statsReader.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			i.termState.SetDocFreq(int(docFreq))
			//if (DEBUG) System.out.println("    dF=" + state.docFreq);
			if i.ite.fr.fieldInfo.GetIndexOptions() == document.INDEX_OPTIONS_DOCS {
				i.termState.SetTotalTermFreq(i.termState.GetDocFreq()) // all postings have freq=1
			} else {
				n, err := i.statsReader.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				i.termState.SetTotalTermFreq(i.termState.GetDocFreq() + int(n))
				//if (DEBUG) System.out.println("    totTF=" + state.totalTermFreq);
			}
		}

		// metadata
		err := i.ite.fr.parent.postingsReader.DecodeTerm(ctx, i.bytesReader, i.ite.fr.fieldInfo, i.termState, absolute)
		if err != nil {
			return err
		}

		i.metaDataUpto++
		absolute = false
	}

	i.termState.SetTermBlockOrd(i.metaDataUpto)

	return nil
}
