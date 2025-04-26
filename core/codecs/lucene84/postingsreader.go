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

var _ types.PostingsReader = &PostingsReader{}

type PostingsReader struct {
	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	version int
}

func NewPostingsReader(ctx context.Context, state *index.SegmentReadState) (*PostingsReader, error) {
	var docIn, posIn, payIn store.IndexInput

	// NOTE: these data files are too costly to verify checksum against all the bytes on open,
	// but for now we at least verify proper structure of the checksum footer: which looks
	// for FOOTER_MAGIC + algorithmID. This is cheap and can detect some forms of corruption
	// such as file truncation.

	docName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, DOC_EXTENSION)

	var err error
	docIn, err = state.Directory.OpenInput(ctx, docName)
	if err != nil {
		return nil, err
	}
	version, err := codecs.CheckIndexHeader(ctx, docIn, DOC_CODEC, VERSION_START, VERSION_CURRENT, state.SegmentInfo.GetID(), state.SegmentSuffix)
	if err != nil {
		return nil, err
	}
	if _, err := codecs.RetrieveChecksum(ctx, docIn); err != nil {
		return nil, err
	}

	if state.FieldInfos.HasProx() {
		proxName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, POS_EXTENSION)

		posIn, err = state.Directory.OpenInput(ctx, proxName)
		if err != nil {
			return nil, err
		}
		if _, err := codecs.CheckIndexHeader(ctx, posIn, POS_CODEC, version, version, state.SegmentInfo.GetID(), state.SegmentSuffix); err != nil {
			return nil, err
		}
		if _, err := codecs.RetrieveChecksum(ctx, posIn); err != nil {
			return nil, err
		}

		if state.FieldInfos.HasPayloads() || state.FieldInfos.HasOffsets() {
			payName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, PAY_EXTENSION)
			payIn, err = state.Directory.OpenInput(ctx, payName)
			if err != nil {
				return nil, err
			}
			if _, err := codecs.CheckIndexHeader(ctx, payIn, PAY_CODEC, version, version, state.SegmentInfo.GetID(), state.SegmentSuffix); err != nil {
				return nil, err
			}
			if _, err := codecs.RetrieveChecksum(ctx, payIn); err != nil {
				return nil, err
			}
		}
	}

	return &PostingsReader{
		docIn: docIn, posIn: posIn, payIn: payIn, version: version,
	}, nil
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

func (p *PostingsReader) NewTermState() (types.BlockTermState, error) {
	return NewIntBlockTermState(), nil
}

func (p *PostingsReader) DecodeTerm(ctx context.Context, in store.DataInput, fieldInfo *document.FieldInfo,
	state types.BlockTermState, absolute bool) error {
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

func (p *PostingsReader) Postings(ctx context.Context, fieldInfo *document.FieldInfo, termState types.BlockTermState,
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

	var everythingEnum *EverythingEnum
	if reuseEnum, ok := reuse.(*EverythingEnum); ok {
		everythingEnum = reuseEnum
		if !everythingEnum.canReuse(p.docIn, fieldInfo) {
			enum, err := p.NewEverythingEnum(fieldInfo)
			if err != nil {
				return nil, err
			}
			everythingEnum = enum
		}
	} else {
		enum, err := p.NewEverythingEnum(fieldInfo)
		if err != nil {
			return nil, err
		}
		everythingEnum = enum
	}
	return everythingEnum.reset(termState.(*IntBlockTermState), flags)
}

func (p *PostingsReader) Impacts(ctx context.Context, fieldInfo *document.FieldInfo,
	state types.BlockTermState, flags int) (index.ImpactsEnum, error) {
	if state.GetDocFreq() <= BLOCK_SIZE {
		enum, err := p.Postings(ctx, fieldInfo, state, nil, flags)
		if err != nil {
			return nil, err
		}

		return coreIndex.NewSlowImpactsEnum(enum), nil
	}
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
