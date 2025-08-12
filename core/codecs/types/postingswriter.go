package types

import (
	"context"
	"errors"
	"io"

	//"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

// PostingsWriter
// TODO: find a better name; this defines the API that the
// terms dict impls use to talk to a postings impl.
// TermsDict + PostingsReader/WriterBase == FieldsProducer/Consumer
type PostingsWriter interface {
	io.Closer

	// Init
	// Called once after startup, before any terms have been added.
	// Implementations typically write a header to the provided termsOut.
	Init(ctx context.Context, termsOut store.IndexOutput, state *index.SegmentWriteState) error

	// WriteTerm
	// Write all postings for one term; use the provided TermsEnum to pull a
	// org. apache. lucene. index. PostingsEnum. This method should not re-position the TermsEnum!
	// It is already positioned on the term that should be written. This method must set the bit
	// in the provided FixedBitSet for every docID written. If no docs were written, this method
	// should return null, and the terms dict will skip the
	WriteTerm(ctx context.Context, term []byte, termsEnum index.TermsEnum, docsSeen index.Bits, norms index.NormsProducer) (BlockTermState, error)

	EncodeTerm(ctx context.Context, out store.DataOutput, fieldInfo *document.FieldInfo, state BlockTermState, absolute bool) error

	SetField(fieldInfo *document.FieldInfo)
}

type PushPostingsWriter interface {
	PostingsWriter

	// NewTermState Return a newly created empty TermState
	NewTermState() (BlockTermState, error)

	// StartTerm Start a new term. Note that a matching call to finishTerm(BlockTermState) is done,
	// only if the term has at least one document.
	StartTerm(norms index.NumericDocValues) error

	// FinishTerm Finishes the current term. The provided BlockTermState contains the term's summary
	// statistics, and will holds metadata from PBF when returned
	FinishTerm(ctx context.Context, state BlockTermState) error

	// StartDoc Adds a new doc in this term. freq will be -1 when term frequencies are omitted for the field.
	StartDoc(docID, freq int) error

	// AddPosition Add a new position and payload, and start/ end offset. A null payload means no payload;
	// a non-null payload with zero length also means no payload. Caller may reuse the BytesRef for the
	// payload between calls (method must fully consume the payload). startOffset and endOffset will be
	// -1 when offsets are not indexed.
	AddPosition(ctx context.Context, position int, payload []byte, startOffset, endOffset int) error

	// FinishDoc Called when we are done adding positions and payloads for each doc.
	FinishDoc() error
}

type PushPostingsWriterBase struct {
	// Reused in writeTerm
	PostingsEnum index.PostingsEnum
	EnumFlags    int

	// of current field being written.
	FieldInfo *document.FieldInfo

	// of current field being written
	IndexOptions document.IndexOptions

	// True if the current field writes freqs.
	WriteFreqs bool

	// True if the current field writes positions.
	WritePositions bool

	// True if the current field writes payloads.
	WritePayloads bool

	// True if the current field writes offsets.
	WriteOffsets bool
}

func (p *PushPostingsWriterBase) SetField(fieldInfo *document.FieldInfo) {
	p.FieldInfo = fieldInfo
	p.IndexOptions = fieldInfo.GetIndexOptions()

	p.WriteFreqs = p.IndexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS
	p.WritePositions = p.IndexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	p.WriteOffsets = p.IndexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	p.WritePayloads = fieldInfo.HasPayloads()

	if p.WriteFreqs == false {
		p.EnumFlags = 0
	} else if p.WritePositions == false {
		p.EnumFlags = index.POSTINGS_ENUM_FREQS
	} else if p.WriteOffsets == false {
		if p.WritePayloads {
			p.EnumFlags = index.POSTINGS_ENUM_PAYLOADS
		} else {
			p.EnumFlags = index.POSTINGS_ENUM_POSITIONS
		}
	} else {
		if p.WritePayloads {
			p.EnumFlags = index.POSTINGS_ENUM_PAYLOADS | index.POSTINGS_ENUM_OFFSETS
		} else {
			p.EnumFlags = index.POSTINGS_ENUM_OFFSETS
		}
	}
}

type WriteTermOptions struct {
	StartTerm    func(norms index.NumericDocValues) error
	StartDoc     func(docID, freq int) error
	AddPosition  func(ctx context.Context, position int, payload []byte, startOffset, endOffset int) error
	FinishDoc    func() error
	NewTermState func() (BlockTermState, error)
	FinishTerm   func(ctx context.Context, state BlockTermState) error
}

func (p *PushPostingsWriterBase) WriteTerm(ctx context.Context, term []byte, termsEnum index.TermsEnum, docsSeen index.Bits,
	norms index.NormsProducer, options *WriteTermOptions) (BlockTermState, error) {

	var normValues index.NumericDocValues
	var err error
	if !p.FieldInfo.HasNorms() {
		normValues, err = norms.GetNorms(p.FieldInfo)
		if err != nil {
			return nil, err
		}
	}

	if err := options.StartTerm(normValues); err != nil {
		return nil, err
	}
	p.PostingsEnum, err = termsEnum.Postings(p.PostingsEnum, p.EnumFlags)
	if err != nil {
		return nil, err
	}

	docFreq := 0
	totalTermFreq := 0
	for {
		docID, err := p.PostingsEnum.NextDoc(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
		}

		docFreq++
		docsSeen.Test(uint(docID))
		freq := -1
		if p.WriteFreqs {
			freq, err = p.PostingsEnum.Freq()
			if err != nil {
				return nil, err
			}
			totalTermFreq += freq
		}
		if err := options.StartDoc(docID, freq); err != nil {
			return nil, err
		}

		if p.WritePositions {
			for i := 0; i < freq; i++ {
				pos, err := p.PostingsEnum.NextPosition()
				if err != nil {
					return nil, err
				}
				var payload []byte
				if p.WritePayloads {
					payload, err = p.PostingsEnum.GetPayload()
					if err != nil {
						return nil, err
					}
				}

				var startOffset, endOffset int
				if p.WriteOffsets {
					startOffset, err = p.PostingsEnum.StartOffset()
					if err != nil {
						return nil, err
					}
					endOffset, err = p.PostingsEnum.EndOffset()
					if err != nil {
						return nil, err
					}
				} else {
					startOffset = -1
					endOffset = -1
				}
				if err := options.AddPosition(ctx, pos, payload, startOffset, endOffset); err != nil {
					return nil, err
				}
			}
		}

		if err := options.FinishDoc(); err != nil {
			return nil, err
		}
	}

	if docFreq == 0 {
		return nil, nil
	} else {
		state, err := options.NewTermState()
		if err != nil {
			return nil, err
		}
		state.SetDocFreq(docFreq)
		state.SetTotalTermFreq(-1)
		if p.WriteFreqs {
			state.SetTotalTermFreq(totalTermFreq)
		}
		if err := options.FinishTerm(ctx, state); err != nil {
			return nil, err
		}
		return state, nil
	}
}
