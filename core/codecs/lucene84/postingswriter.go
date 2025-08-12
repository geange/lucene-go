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

var _ types.PushPostingsWriter = &PostingsWriter{}

type PostingsWriter struct {
	*types.PushPostingsWriterBase

	docOut store.IndexOutput
	posOut store.IndexOutput
	payOut store.IndexOutput

	emptyState *IntBlockTermState
	lastState  *IntBlockTermState

	// Holds starting file pointers for current term:
	docStartFP int64
	posStartFP int64
	payStartFP int64

	docDeltaBuffer []uint64
	freqBuffer     []uint64
	docBufferUpto  int

	posDeltaBuffer         []uint64
	payloadLengthBuffer    []uint64
	offsetStartDeltaBuffer []uint64
	offsetLengthBuffer     []uint64
	posBufferUpto          int

	payloadBytes []byte
	//payloadByteUpto          int
	lastBlockDocID           int
	lastBlockPosFP           int64
	lastBlockPayFP           int64
	lastBlockPosBufferUpto   int
	lastBlockPayloadByteUpto int

	lastDocID       int
	lastPosition    int
	lastStartOffset int
	docCount        int

	pforUtil     *PForUtil
	forDeltaUtil *ForDeltaUtil
	skipWriter   SkipWriter

	fieldHasNorms                  bool
	norms                          index.NumericDocValues
	competitiveFreqNormAccumulator coreIndex.CompetitiveImpactAccumulator
}

func (p *PostingsWriter) Close() error {
	ctx := context.Background()
	if p.docOut != nil {
		if err := codecs.WriteFooter(ctx, p.docOut); err != nil {
			return err
		}
	}
	if p.posOut != nil {
		if err := codecs.WriteFooter(ctx, p.posOut); err != nil {
			return err
		}
	}
	if p.payOut != nil {
		if err := codecs.WriteFooter(ctx, p.payOut); err != nil {
			return err
		}
	}

	if p.docOut != nil {
		if err := p.docOut.Close(); err != nil {
			return err
		}
	}
	if p.posOut != nil {
		if err := p.posOut.Close(); err != nil {
			return err
		}
	}
	if p.payOut != nil {
		if err := p.payOut.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostingsWriter) Init(ctx context.Context, termsOut store.IndexOutput, state *index.SegmentWriteState) error {
	err := codecs.WriteIndexHeader(ctx, termsOut, TERMS_CODEC, VERSION_CURRENT, state.SegmentInfo.GetID(), state.SegmentSuffix)
	if err != nil {
		return err
	}
	return termsOut.WriteUvarint(ctx, uint64(BLOCK_SIZE))
}

func (p *PostingsWriter) WriteTerm(ctx context.Context, term []byte, termsEnum index.TermsEnum,
	docsSeen index.Bits, norms index.NormsProducer) (types.BlockTermState, error) {

	return p.PushPostingsWriterBase.WriteTerm(ctx, term, termsEnum, docsSeen, norms, &types.WriteTermOptions{
		StartTerm:    p.StartTerm,
		StartDoc:     p.StartDoc,
		AddPosition:  p.AddPosition,
		FinishDoc:    p.FinishDoc,
		NewTermState: p.NewTermState,
		FinishTerm:   p.FinishTerm,
	})
}

func (p *PostingsWriter) EncodeTerm(ctx context.Context, out store.DataOutput, fieldInfo *document.FieldInfo, _state types.BlockTermState, absolute bool) error {
	state, ok := _state.(*IntBlockTermState)
	if !ok {
		return errors.New("state is not *IntBlockTermState")
	}
	if absolute {
		p.lastState = NewIntBlockTermState()
	}

	if p.lastState.SingletonDocID != -1 && state.SingletonDocID != -1 && state.DocStartFP == p.lastState.DocStartFP {
		// With runs of rare values such as ID fields, the increment of pointers in the docs file is often 0.
		// Furthermore some ID schemes like auto-increment IDs or Flake IDs are monotonic, so we encode the delta
		// between consecutive doc IDs to save space.
		delta := state.SingletonDocID - p.lastState.SingletonDocID
		num := zigzag.Encode(int64(delta))<<1 | 0x01
		if err := out.WriteUvarint(ctx, num); err != nil {
			return err
		}
	} else {
		if err := out.WriteUvarint(ctx, uint64(state.DocStartFP-p.lastState.DocStartFP)<<1); err != nil {
			return err
		}
		if state.SingletonDocID != -1 {
			if err := out.WriteUvarint(ctx, uint64(state.SingletonDocID)); err != nil {
				return err
			}
		}
	}

	if p.WritePositions {
		if err := out.WriteUvarint(ctx, uint64(state.PosStartFP-p.lastState.PosStartFP)); err != nil {
			return err
		}
		if p.WritePayloads || p.WriteOffsets {
			if err := out.WriteUvarint(ctx, uint64(state.PayStartFP-p.lastState.PayStartFP)); err != nil {
				return err
			}
		}
	}
	if p.WritePositions {
		if state.LastPosBlockOffset != -1 {
			if err := out.WriteUvarint(ctx, uint64(state.LastPosBlockOffset)); err != nil {
				return err
			}
		}
	}
	if state.SkipOffset != -1 {
		if err := out.WriteUvarint(ctx, uint64(state.SkipOffset)); err != nil {
			return err
		}
	}
	p.lastState = state
	return nil
}

func (p *PostingsWriter) NewTermState() (types.BlockTermState, error) {
	return NewIntBlockTermState(), nil
}

func (p *PostingsWriter) StartTerm(norms index.NumericDocValues) error {
	p.docStartFP = p.docOut.GetFilePointer()
	if p.WritePositions {
		p.posStartFP = p.posOut.GetFilePointer()
		if p.WritePayloads || p.WriteOffsets {
			p.payStartFP = p.payOut.GetFilePointer()
		}
	}
	p.lastDocID = 0
	p.lastBlockDocID = -1
	if err := p.skipWriter.ResetSkip(); err != nil {
		return err
	}
	p.norms = norms
	p.competitiveFreqNormAccumulator.Clear()
	return nil
}

func (p *PostingsWriter) FinishTerm(ctx context.Context, _state types.BlockTermState) error {
	state := _state.(*IntBlockTermState)

	// TODO: wasteful we are counting this (counting # docs
	// for this term) in two places?

	// docFreq == 1, don't write the single docid/freq to a separate file along with a pointer to it.
	var singletonDocID int64
	if state.DocFreq == 1 {
		// pulse the singleton docid into the term dictionary, freq is implicitly totalTermFreq
		singletonDocID = int64(p.docDeltaBuffer[0])
	} else {
		singletonDocID = -1
		// vInt encode the remaining doc deltas and freqs:
		for i := 0; i < p.docBufferUpto; i++ {
			docDelta := p.docDeltaBuffer[i]
			freq := p.freqBuffer[i]
			if !p.WriteFreqs {
				if err := p.docOut.WriteUvarint(ctx, docDelta); err != nil {
					return err
				}
			} else if freq == 1 {
				if err := p.docOut.WriteUvarint(ctx, (docDelta<<1)|1); err != nil {
					return err
				}
			} else {
				if err := p.docOut.WriteUvarint(ctx, docDelta<<1); err != nil {
					return err
				}
				if err := p.docOut.WriteUvarint(ctx, freq); err != nil {
					return err
				}
			}
		}
	}

	var lastPosBlockOffset int64

	if p.WritePositions {
		// totalTermFreq is just total number of positions(or payloads, or offsets)
		// associated with current term.
		if state.TotalTermFreq > BLOCK_SIZE {
			// record file offset for last pos in last block
			lastPosBlockOffset = p.posOut.GetFilePointer() - p.posStartFP
		} else {
			lastPosBlockOffset = -1
		}
		if p.posBufferUpto > 0 {
			// TODO: should we send offsets/payloads to
			// .pay...?  seems wasteful (have to store extra
			// vLong for low (< BLOCK_SIZE) DF terms = vast vast
			// majority)

			// vInt encode the remaining positions/payloads/offsets:
			lastPayloadLength := int64(-1) // force first payload length to be written
			lastOffsetLength := int64(-1)  // force first offset length to be written
			payloadBytesReadUpto := int64(0)
			for i := 0; i < p.posBufferUpto; i++ {
				posDelta := p.posDeltaBuffer[i]
				if p.WritePayloads {
					payloadLength := int64(p.payloadLengthBuffer[i])
					if payloadLength != lastPayloadLength {
						lastPayloadLength = payloadLength
						if err := p.posOut.WriteUvarint(ctx, (posDelta<<1)|1); err != nil {
							return err
						}
						if err := p.posOut.WriteUvarint(ctx, uint64(payloadLength)); err != nil {
							return err
						}
					} else {
						if err := p.posOut.WriteUvarint(ctx, posDelta<<1); err != nil {
							return err
						}
					}

					if payloadLength != 0 {
						if _, err := p.posOut.Write(p.payloadBytes[payloadBytesReadUpto : payloadBytesReadUpto+payloadLength]); err != nil {
							return err
						}
						payloadBytesReadUpto += payloadLength
					}
				} else {
					if err := p.posOut.WriteUvarint(ctx, posDelta); err != nil {
						return err
					}
				}

				if p.WriteOffsets {
					delta := p.offsetStartDeltaBuffer[i]
					length := int64(p.offsetLengthBuffer[i])
					if length == (lastOffsetLength) {
						if err := p.posOut.WriteUvarint(ctx, delta<<1); err != nil {
							return err
						}
					} else {
						if err := p.posOut.WriteUvarint(ctx, delta<<1|1); err != nil {
							return err
						}
						if err := p.posOut.WriteUvarint(ctx, uint64(length)); err != nil {
							return err
						}
						lastOffsetLength = length
					}
				}
			}

			if p.WritePayloads {
				p.payloadBytes = p.payloadBytes[:0]
			}
		}
	} else {
		lastPosBlockOffset = -1
	}

	var skipOffset int64
	if p.docCount > BLOCK_SIZE {
		skip, err := p.skipWriter.WriteSkip(ctx, p.docOut)
		if err != nil {
			return err
		}
		skipOffset = skip - p.docStartFP
	} else {
		skipOffset = -1
	}

	state.DocStartFP = p.docStartFP
	state.PosStartFP = p.posStartFP
	state.PayStartFP = p.payStartFP
	state.SingletonDocID = int(singletonDocID)
	state.SkipOffset = skipOffset
	state.LastPosBlockOffset = lastPosBlockOffset
	p.docBufferUpto = 0
	p.posBufferUpto = 0
	p.lastDocID = 0
	p.docCount = 0
	return nil
}

func (p *PostingsWriter) StartDoc(docID, freq int) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) AddPosition(ctx context.Context, position int, payload []byte, startOffset, endOffset int) error {
	if position > coreIndex.MAX_POSITION {
		return errors.New("position=" + fmt.Sprint(position) + " is too large (> IndexWriter.MAX_POSITION=" + fmt.Sprint(coreIndex.MAX_POSITION) + ")")
	}
	if position < 0 {
		return errors.New("position=" + fmt.Sprint(position) + " is < 0")
	}
	p.posDeltaBuffer[p.posBufferUpto] = uint64(position - p.lastPosition)
	if p.WritePayloads {
		if len(payload) == 0 {
			// no payload
			p.payloadLengthBuffer[p.posBufferUpto] = 0
		} else {
			p.payloadLengthBuffer[p.posBufferUpto] = uint64(len(payload))
			p.payloadBytes = append(p.payloadBytes, payload...)
		}
	}

	if p.WriteOffsets {
		p.offsetStartDeltaBuffer[p.posBufferUpto] = uint64(startOffset - p.lastStartOffset)
		p.offsetLengthBuffer[p.posBufferUpto] = uint64(endOffset - startOffset)
		p.lastStartOffset = startOffset
	}

	p.posBufferUpto++
	p.lastPosition = position
	if p.posBufferUpto == BLOCK_SIZE {
		if err := p.pforUtil.Encode(ctx, p.posDeltaBuffer, p.posOut); err != nil {
			return err
		}

		if p.WritePayloads {
			if err := p.pforUtil.Encode(ctx, p.payloadLengthBuffer, p.payOut); err != nil {
				return err
			}
			if err := p.payOut.WriteUvarint(ctx, uint64(len(p.payloadBytes))); err != nil {
				return err
			}
			if _, err := p.payOut.Write(p.payloadBytes); err != nil {
				return err
			}
			p.payloadBytes = p.payloadBytes[:0]
		}
		if p.WriteOffsets {
			if err := p.pforUtil.Encode(ctx, p.offsetStartDeltaBuffer, p.payOut); err != nil {
				return err
			}
			if err := p.pforUtil.Encode(ctx, p.offsetLengthBuffer, p.payOut); err != nil {
				return err
			}
		}
		p.posBufferUpto = 0
	}
	return nil
}

func (p *PostingsWriter) FinishDoc() error {
	// Since we don't know df for current term, we had to buffer
	// those skip data for each block, and when a new doc comes,
	// write them to skip file.
	if p.docBufferUpto == BLOCK_SIZE {
		p.lastBlockDocID = p.lastDocID
		if p.posOut != nil {
			if p.payOut != nil {
				p.lastBlockPayFP = p.payOut.GetFilePointer()
			}
			p.lastBlockPosFP = p.posOut.GetFilePointer()
			p.lastBlockPosBufferUpto = p.posBufferUpto
			p.lastBlockPayloadByteUpto = len(p.payloadBytes)
		}
		p.docBufferUpto = 0
	}
	return nil
}
