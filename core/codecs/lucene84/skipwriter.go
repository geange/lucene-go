package lucene84

import (
	"context"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

type SkipWriter struct {
	sw  *skipWriter
	mwc *coreIndex.MultiLevelSkipListWriterContext
}

func NewSkipWriter(maxSkipLevels, blockSize, docCount int,
	docOut, posOut, payOut store.IndexOutput) *SkipWriter {
	mwc := coreIndex.NewMultiLevelSkipListWriterContext(blockSize, 8, maxSkipLevels, docCount)
	sw := newSkipWriter(maxSkipLevels,
		docOut, posOut, payOut)
	return &SkipWriter{
		sw:  sw,
		mwc: mwc,
	}
}

func (s *SkipWriter) initSkip() error {
	if !s.sw.initialized {
		s.mwc.ResetSkip()
		arrayFill(s.sw.lastSkipDoc, 0)
		arrayFill(s.sw.lastSkipDocPointer, uint64(s.sw.lastDocFP))

		if s.sw.fieldHasPositions {
			for i := range s.sw.lastSkipPosPointer {
				s.sw.lastSkipPosPointer[i] = uint64(s.sw.lastPosFP)
			}

			if s.sw.fieldHasPayloads {
				arrayFill(s.sw.lastPayloadByteUpto, 0)
			}
			if s.sw.fieldHasOffsets || s.sw.fieldHasPayloads {
				arrayFill(s.sw.lastSkipPayPointer, uint64(s.sw.lastPayFP))
			}
		}
		// sets of competitive freq,norm pairs should be empty at this point

		s.sw.initialized = true
	}
	return nil
}

func (s *SkipWriter) WriteSkip(ctx context.Context, output store.IndexOutput) (int64, error) {
	return s.sw.WriteSkip(ctx, output, s.mwc)
}

func (s *SkipWriter) ResetSkip() error {
	return s.sw.ResetSkip(s.mwc)
}

func (s *SkipWriter) BufferSkip(ctx context.Context, doc int, competitiveFreqNorms *coreIndex.CompetitiveImpactAccumulator,
	numDocs int, posFP, payFP uint64, posBufferUpto, payloadByteUpto int) error {
	if err := s.initSkip(); err != nil {
		return err
	}
	s.sw.curDoc = doc
	s.sw.curDocPointer = uint64(s.sw.docOut.GetFilePointer())
	s.sw.curPosPointer = posFP
	s.sw.curPayPointer = payFP
	s.sw.curPosBufferUpto = posBufferUpto
	s.sw.curPayloadByteUpto = payloadByteUpto
	s.sw.curCompetitiveFreqNorms[0].AddAll(competitiveFreqNorms)
	return s.mwc.BufferSkip(ctx, numDocs, s.sw)
}

var _ coreIndex.MultiLevelSkipListWriterSPI = &skipWriter{}

type skipWriter struct {
	lastSkipDoc         []int
	lastSkipDocPointer  []uint64
	lastSkipPosPointer  []uint64
	lastSkipPayPointer  []uint64
	lastPayloadByteUpto []int

	docOut store.IndexOutput
	posOut store.IndexOutput
	payOut store.IndexOutput

	curDoc                  int
	curDocPointer           uint64
	curPosPointer           uint64
	curPayPointer           uint64
	curPosBufferUpto        int
	curPayloadByteUpto      int
	curCompetitiveFreqNorms []*coreIndex.CompetitiveImpactAccumulator
	fieldHasPositions       bool
	fieldHasOffsets         bool
	fieldHasPayloads        bool

	freqNormOut *store.BufferDataOutput

	// tricky: we only skip data for blocks (terms with more than 128 docs), but re-init'ing the skipper
	// is pretty slow for rare terms in large segments as we have to fill O(log #docs in segment) of junk.
	// this is the vast majority of terms (worst case: ID field or similar).  so in resetSkip() we save
	// away the previous pointers, and lazy-init only if we need to buffer skip data for the term.
	initialized bool
	lastDocFP   int64
	lastPosFP   int64
	lastPayFP   int64
}

func newSkipWriter(maxSkipLevels int,
	docOut, posOut, payOut store.IndexOutput) *skipWriter {

	this := &skipWriter{
		lastSkipDoc:        make([]int, maxSkipLevels),
		lastSkipDocPointer: make([]uint64, maxSkipLevels),
		docOut:             docOut,
		posOut:             posOut,
		payOut:             payOut,
	}

	if posOut != nil {
		this.lastSkipPosPointer = make([]uint64, maxSkipLevels)
		if payOut != nil {
			this.lastSkipPayPointer = make([]uint64, maxSkipLevels)
		}
		this.lastPayloadByteUpto = make([]int, maxSkipLevels)
	}
	this.curCompetitiveFreqNorms = make([]*coreIndex.CompetitiveImpactAccumulator, maxSkipLevels)
	for i := 0; i < maxSkipLevels; i++ {
		this.curCompetitiveFreqNorms[i] = coreIndex.NewCompetitiveImpactAccumulator()
	}
	return this
}

func arrayFill[T any](array []T, value T) {
	for i := range array {
		array[i] = value
	}
}

func (s *skipWriter) ResetSkip(mwc *coreIndex.MultiLevelSkipListWriterContext) error {
	s.lastDocFP = s.docOut.GetFilePointer()
	if s.fieldHasPositions {
		s.lastPosFP = s.posOut.GetFilePointer()
		if s.fieldHasOffsets || s.fieldHasPayloads {
			s.lastPayFP = s.payOut.GetFilePointer()
		}
	}
	if s.initialized {
		for _, acc := range s.curCompetitiveFreqNorms {
			acc.Clear()
		}
	}
	s.initialized = false
	return nil
}

func (s *skipWriter) WriteSkipData(ctx context.Context, level int, skipBuffer store.IndexOutput,
	mwc *coreIndex.MultiLevelSkipListWriterContext) error {

	delta := s.curDoc - s.lastSkipDoc[level]

	if err := skipBuffer.WriteUvarint(ctx, uint64(delta)); err != nil {
		return err
	}
	s.lastSkipDoc[level] = s.curDoc

	if err := skipBuffer.WriteUvarint(ctx, s.curDocPointer-s.lastSkipDocPointer[level]); err != nil {
		return err
	}
	s.lastSkipDocPointer[level] = s.curDocPointer

	if s.fieldHasPositions {

		if err := skipBuffer.WriteUvarint(ctx, s.curPosPointer-s.lastSkipPosPointer[level]); err != nil {
			return err
		}
		s.lastSkipPosPointer[level] = s.curPosPointer
		if err := skipBuffer.WriteUvarint(ctx, uint64(s.curPosBufferUpto)); err != nil {
			return err
		}

		if s.fieldHasPayloads {
			if err := skipBuffer.WriteUvarint(ctx, uint64(s.curPayloadByteUpto)); err != nil {
				return err
			}
		}

		if s.fieldHasOffsets || s.fieldHasPayloads {
			if err := skipBuffer.WriteUvarint(ctx, uint64(s.curPayPointer)-s.lastSkipPayPointer[level]); err != nil {
				return err
			}
			s.lastSkipPayPointer[level] = s.curPayPointer
		}
	}

	competitiveFreqNorms := s.curCompetitiveFreqNorms[level]
	if level+1 < mwc.NumberOfSkipLevels {
		s.curCompetitiveFreqNorms[level+1].AddAll(competitiveFreqNorms)
	}
	if err := writeImpacts(ctx, competitiveFreqNorms, s.freqNormOut); err != nil {
		return err
	}
	if err := skipBuffer.WriteUvarint(ctx, uint64(s.freqNormOut.Size())); err != nil {
		return err
	}
	//copy(skipBuffer, s.freqNormOut)
	if err := s.freqNormOut.CopyTo(skipBuffer); err != nil {
		return err
	}
	s.freqNormOut.Reset()
	competitiveFreqNorms.Clear()
	return nil
}

func (s *skipWriter) WriteSkip(ctx context.Context, output store.IndexOutput, mwc *coreIndex.MultiLevelSkipListWriterContext) (int64, error) {
	skipPointer := output.GetFilePointer()
	//System.out.println("skipper.writeSkip fp=" + skipPointer);
	if mwc.SkipBuffer == nil || len(mwc.SkipBuffer) == 0 {
		return skipPointer, nil
	}

	for level := mwc.NumberOfSkipLevels - 1; level > 0; level-- {
		buffer := mwc.SkipBuffer[level]

		length := buffer.GetFilePointer()
		if length > 0 {
			if err := s.WriteLevelLength(ctx, length, output); err != nil {
				return 0, err
			}
			if _, err := output.Write(buffer.Bytes()); err != nil {
				return 0, err
			}
		}
	}
	if _, err := output.Write(mwc.SkipBuffer[0].Bytes()); err != nil {
		return 0, err
	}

	return skipPointer, nil
}

func (s *skipWriter) WriteLevelLength(ctx context.Context, levelLength int64, output store.IndexOutput) error {
	return output.WriteUvarint(ctx, uint64(levelLength))
}

func (s *skipWriter) WriteChildPointer(ctx context.Context, childPointer int64, skipBuffer store.DataOutput) error {
	return skipBuffer.WriteUvarint(ctx, uint64(childPointer))
}

func writeImpacts(ctx context.Context, acc *coreIndex.CompetitiveImpactAccumulator, out store.DataOutput) error {
	impacts := acc.GetCompetitiveFreqNormPairs()
	previous := coreIndex.NewImpact(0, 0)

	for _, impact := range impacts {
		freqDelta := uint64(impact.GetFreq() - previous.GetFreq() - 1)
		normDelta := impact.GetNorm() - previous.GetNorm() - 1
		if normDelta == 0 {
			// most of time, norm only increases by 1, so we can fold everything in a single byte
			if err := out.WriteUvarint(ctx, freqDelta<<1); err != nil {
				return err
			}
		} else {
			if err := out.WriteUvarint(ctx, (freqDelta<<1)|1); err != nil {
				return err
			}
			if err := out.WriteZInt64(ctx, normDelta); err != nil {
				return err
			}
		}
		previous = impact
	}
	return nil
}
