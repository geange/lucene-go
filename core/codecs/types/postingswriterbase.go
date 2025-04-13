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
	Init(termsOut store.IndexOutput, state *index.SegmentWriteState) error

	// WriteTerm
	// Write all postings for one term; use the provided TermsEnum to pull a
	// org. apache. lucene. index. PostingsEnum. This method should not re-position the TermsEnum!
	// It is already positioned on the term that should be written. This method must set the bit
	// in the provided FixedBitSet for every docID written. If no docs were written, this method
	// should return null, and the terms dict will skip the
	WriteTerm(term []byte, termsEnum index.TermsEnum, docsSeen index.Bits, norms index.NormsProducer) (*BlockTermState, error)

	EncodeTerm(out store.DataOutput, fieldInfo *document.FieldInfo, state *BlockTermState, absolute bool) error

	SetField(fieldInfo *document.FieldInfo)
}

type PushPostingsWriter interface {
	PostingsWriter

	// NewTermState Return a newly created empty TermState
	NewTermState() (*BlockTermState, error)

	// StartTerm Start a new term. Note that a matching call to finishTerm(BlockTermState) is done,
	// only if the term has at least one document.
	StartTerm(norms index.NumericDocValues) error

	// FinishTerm Finishes the current term. The provided BlockTermState contains the term's summary
	// statistics, and will holds metadata from PBF when returned
	FinishTerm(state *BlockTermState) error

	// StartDoc Adds a new doc in this term. freq will be -1 when term frequencies are omitted for the field.
	StartDoc(docID, freq int) error

	// AddPosition Add a new position and payload, and start/ end offset. A null payload means no payload;
	// a non-null payload with zero length also means no payload. Caller may reuse the BytesRef for the
	// payload between calls (method must fully consume the payload). startOffset and endOffset will be
	// -1 when offsets are not indexed.
	AddPosition(position int, payload []byte, startOffset, endOffset int) error

	// FinishDoc Called when we are done adding positions and payloads for each doc.
	FinishDoc() error
}

type PushPostingsWriterBase struct {
	// Reused in writeTerm
	postingsEnum index.PostingsEnum
	enumFlags    int

	// of current field being written.
	fieldInfo *document.FieldInfo

	// of current field being written
	indexOptions document.IndexOptions

	// True if the current field writes freqs.
	writeFreqs bool

	// True if the current field writes positions.
	writePositions bool

	// True if the current field writes payloads.
	writePayloads bool

	// True if the current field writes offsets.
	writeOffsets bool
}

func (p *PushPostingsWriterBase) SetField(fieldInfo *document.FieldInfo) {
	p.fieldInfo = fieldInfo
	p.indexOptions = fieldInfo.GetIndexOptions()

	p.writeFreqs = p.indexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS
	p.writePositions = p.indexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	p.writeOffsets = p.indexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	p.writePayloads = fieldInfo.HasPayloads()

	if p.writeFreqs == false {
		p.enumFlags = 0
	} else if p.writePositions == false {
		p.enumFlags = index.POSTINGS_ENUM_FREQS
	} else if p.writeOffsets == false {
		if p.writePayloads {
			p.enumFlags = index.POSTINGS_ENUM_PAYLOADS
		} else {
			p.enumFlags = index.POSTINGS_ENUM_POSITIONS
		}
	} else {
		if p.writePayloads {
			p.enumFlags = index.POSTINGS_ENUM_PAYLOADS | index.POSTINGS_ENUM_OFFSETS
		} else {
			p.enumFlags = index.POSTINGS_ENUM_OFFSETS
		}
	}
}

type WriteTermOptions struct {
	StartTerm    func(norms index.NumericDocValues) error
	StartDoc     func(docID, freq int) error
	AddPosition  func(position int, payload []byte, startOffset, endOffset int) error
	FinishDoc    func() error
	NewTermState func() (*BlockTermState, error)
	FinishTerm   func(state *BlockTermState) error
}

func (p *PushPostingsWriterBase) WriteTerm(ctx context.Context, term []byte, termsEnum index.TermsEnum, docsSeen index.Bits,
	norms index.NormsProducer, options *WriteTermOptions) (*BlockTermState, error) {

	var normValues index.NumericDocValues
	var err error
	if !p.fieldInfo.HasNorms() {
		normValues, err = norms.GetNorms(p.fieldInfo)
		if err != nil {
			return nil, err
		}
	}

	if err := options.StartTerm(normValues); err != nil {
		return nil, err
	}
	p.postingsEnum, err = termsEnum.Postings(p.postingsEnum, p.enumFlags)
	if err != nil {
		return nil, err
	}

	docFreq := 0
	totalTermFreq := 0
	for {
		docID, err := p.postingsEnum.NextDoc(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
		}

		docFreq++
		docsSeen.Test(uint(docID))
		freq := -1
		if p.writeFreqs {
			freq, err = p.postingsEnum.Freq()
			if err != nil {
				return nil, err
			}
			totalTermFreq += freq
		}
		if err := options.StartDoc(docID, freq); err != nil {
			return nil, err
		}

		if p.writePositions {
			for i := 0; i < freq; i++ {
				pos, err := p.postingsEnum.NextPosition()
				if err != nil {
					return nil, err
				}
				var payload []byte
				if p.writePayloads {
					payload, err = p.postingsEnum.GetPayload()
					if err != nil {
						return nil, err
					}
				}

				var startOffset, endOffset int
				if p.writeOffsets {
					startOffset, err = p.postingsEnum.StartOffset()
					if err != nil {
						return nil, err
					}
					endOffset, err = p.postingsEnum.EndOffset()
					if err != nil {
						return nil, err
					}
				} else {
					startOffset = -1
					endOffset = -1
				}
				if err := options.AddPosition(pos, payload, startOffset, endOffset); err != nil {
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
		state.DocFreq = docFreq
		state.TotalTermFreq = -1
		if p.writeFreqs {
			state.TotalTermFreq = totalTermFreq
		}
		if err := options.FinishTerm(state); err != nil {
			return nil, err
		}
		return state, nil
	}
}
