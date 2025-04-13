package lucene84

import (
	"context"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/zigzag"
)

var _ types.PostingsReaderBase = &PostingsReader{}

type PostingsReader struct {
	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	version int
}

func (p *PostingsReader) Init(ctx context.Context, termsIn store.IndexInput, state *index.SegmentReadState) error {
	if _, err := codecs.CheckIndexHeader(ctx, termsIn, TERMS_CODEC, VERSION_START, VERSION_CURRENT,
		state.SegmentInfo.GetID(), state.SegmentSuffix); err != nil {
		return err
	}
	indexBlockSize, err := termsIn.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	if indexBlockSize != BLOCK_SIZE {
		return fmt.Errorf("index-time BLOCK_SIZE ( %d ) != read-time BLOCK_SIZE ( %d )",
			indexBlockSize, BLOCK_SIZE)
	}
	return nil
}

func (p *PostingsReader) NewTermState() (index.TermState, error) {
	return NewIntBlockTermState(), nil
}

func (p *PostingsReader) DecodeTerm(ctx context.Context, in store.DataInput, fieldInfo *document.FieldInfo,
	state index.TermState, absolute bool) error {
	termState, ok := state.(*IntBlockTermState)
	if !ok {
		return errors.New("state is not *IntBlockTermState")
	}

	fieldHasPositions := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	fieldHasOffsets := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	fieldHasPayloads := fieldInfo.HasPayloads()

	if absolute {
		termState.DocStartFP = 0
		termState.PosStartFP = 0
		termState.PayStartFP = 0
	}

	if p.version >= VERSION_COMPRESSED_TERMS_DICT_IDS {
		l, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		if (l & 0x01) == 0 {
			termState.DocStartFP += int64(l >> 1)
			if termState.DocFreq == 1 {
				num, err := in.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				termState.SingletonDocID = int(num)
			} else {
				termState.SingletonDocID = -1
			}
		} else {
			termState.SingletonDocID += int(zigzag.Decode(l >> 1))
		}
	} else {
		num, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		termState.DocStartFP += int64(num)
	}

	if fieldHasPositions {
		num, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		termState.PosStartFP += int64(num)
		if fieldHasOffsets || fieldHasPayloads {
			n, err := in.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			termState.PayStartFP += int64(n)
		}
	}

	if p.version < VERSION_COMPRESSED_TERMS_DICT_IDS {
		if termState.DocFreq == 1 {
			n, err := in.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			termState.SingletonDocID = int(n)
		} else {
			termState.SingletonDocID = -1
		}
	}

	if fieldHasPositions {
		if termState.TotalTermFreq > BLOCK_SIZE {
			n, err := in.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			termState.LastPosBlockOffset = int64(n)
		} else {
			termState.LastPosBlockOffset = -1
		}
	}

	if termState.DocFreq > BLOCK_SIZE {
		n, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		termState.SkipOffset = int64(n)
	} else {
		termState.SkipOffset = -1
	}

	return nil
}

func (p *PostingsReader) Postings(ctx context.Context, fieldInfo *document.FieldInfo, termState index.TermState,
	reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {

	indexHasPositions := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS

	if indexHasPositions == false || !coreIndex.FeatureRequested(flags, coreIndex.POSTINGS_ENUM_POSITIONS) {
		var docsEnum *BlockDocsEnum

		docsEnum, ok := reuse.(*BlockDocsEnum)
		if ok {
			if !docsEnum.CanReuse(p.docIn, fieldInfo) {
				docsEnum = NewBlockDocsEnum(fieldInfo)
			}
		} else {
			docsEnum = NewBlockDocsEnum(fieldInfo)
		}
		return docsEnum.Reset(termState.(*IntBlockTermState), flags)
	}

	panic("every enum")
}

func (p *PostingsReader) Impacts(ctx context.Context, fieldInfo *document.FieldInfo, state index.TermState,
	flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsReader) Close() error {
	if err := p.docIn.Close(); err != nil {
		return err
	}
	if err := p.posIn.Close(); err != nil {
		return err
	}
	if err := p.payIn.Close(); err != nil {
		return err
	}
	return nil
}

func (p *PostingsReader) CheckIntegrity() error {
	if p.docIn != nil {
		if _, err := codecs.ChecksumEntireFile(context.Background(), p.docIn); err != nil {
			return err
		}
	}

	if p.posIn != nil {
		if _, err := codecs.ChecksumEntireFile(context.Background(), p.posIn); err != nil {
			return err
		}
	}

	if p.payIn != nil {
		if _, err := codecs.ChecksumEntireFile(context.Background(), p.payIn); err != nil {
			return err
		}
	}
	return nil
}
