package simpletext

import (
	"context"
	"fmt"

	"github.com/geange/lucene-go/codecs/utils"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

const (
	skipMultiplier = 3
	maxSkipLevels  = 4

	BLOCK_SIZE = 8
)

var (
	SKIP_LIST     = []byte("    skipList ")
	LEVEL_LENGTH  = []byte("      levelLength ")
	LEVEL         = []byte("      level ")
	SKIP_DOC      = []byte("        skipDoc ")
	SKIP_DOC_FP   = []byte("        skipDocFP ")
	IMPACTS       = []byte("        impacts ")
	IMPACT        = []byte("          impact ")
	FREQ          = []byte("            freq ")
	NORM          = []byte("            norm ")
	IMPACTS_END   = []byte("        impactsEnd ")
	CHILD_POINTER = []byte("        childPointer ")
)

type SkipWriter struct {
	sw  *skipWriter
	mwc *coreIndex.MultiLevelSkipListWriterContext
}

var _ coreIndex.MultiLevelSkipListWriterSPI = &skipWriter{}

type skipWriter struct {
	wroteHeaderPerLevelMap  map[int]bool
	curDoc                  int
	curDocFilePointer       int64
	curCompetitiveFreqNorms []*coreIndex.CompetitiveImpactAccumulator
}

func NewSkipWriter(writeState *index.SegmentWriteState) (*SkipWriter, error) {

	writer := &skipWriter{
		wroteHeaderPerLevelMap:  make(map[int]bool),
		curDoc:                  0,
		curDocFilePointer:       0,
		curCompetitiveFreqNorms: make([]*coreIndex.CompetitiveImpactAccumulator, maxSkipLevels),
	}
	for i := range writer.curCompetitiveFreqNorms {
		writer.curCompetitiveFreqNorms[i] = coreIndex.NewCompetitiveImpactAccumulator()
	}

	maxDoc, err := writeState.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}
	mwc := coreIndex.NewMultiLevelSkipListWriterContext(BLOCK_SIZE, skipMultiplier, maxSkipLevels, maxDoc)

	err = writer.ResetSkip(mwc)
	if err != nil {
		return nil, err
	}

	return &SkipWriter{
		sw:  writer,
		mwc: mwc,
	}, nil
}

func (s *skipWriter) WriteSkipData(ctx context.Context, level int,
	skipBuffer store.IndexOutput, mwc *coreIndex.MultiLevelSkipListWriterContext) error {

	w := utils.NewTextWriter(skipBuffer)

	// 检查是否已经写入过header，每个level只写一次，
	if !s.wroteHeaderPerLevelMap[level] {
		if err := w.Bytes(LEVEL); err != nil {
			return err
		}
		if err := w.String(fmt.Sprint(level)); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		s.wroteHeaderPerLevelMap[level] = true
	}
	if err := w.Bytes(SKIP_DOC); err != nil {
		return err
	}
	if err := w.String(fmt.Sprint(s.curDoc)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	if err := w.Bytes(SKIP_DOC_FP); err != nil {
		return err
	}
	if err := w.String(fmt.Sprint(s.curDocFilePointer)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	competitiveFreqNorms := s.curCompetitiveFreqNorms[level]
	impacts := competitiveFreqNorms.GetCompetitiveFreqNormPairs()
	//assert impacts.size() > 0;
	if level+1 < mwc.NumberOfSkipLevels {
		s.curCompetitiveFreqNorms[level+1].AddAll(competitiveFreqNorms)
	}
	if err := w.Bytes(IMPACTS); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}
	for _, impact := range impacts {
		if err := w.Bytes(IMPACT); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		if err := w.Bytes(FREQ); err != nil {
			return err
		}
		if err := w.String(fmt.Sprint(impact.GetFreq())); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		if err := w.Bytes(NORM); err != nil {
			return err
		}
		if err := w.String(fmt.Sprint(impact.GetNorm())); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}
	}
	if err := w.Bytes(IMPACTS_END); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}
	competitiveFreqNorms.Clear()

	return nil
}

func (s *skipWriter) ResetSkip(mwc *coreIndex.MultiLevelSkipListWriterContext) error {
	mwc.ResetSkip()

	clear(s.wroteHeaderPerLevelMap)
	s.curDoc = -1
	s.curDocFilePointer = -1
	for _, norm := range s.curCompetitiveFreqNorms {
		norm.Clear()
	}
	return nil
}

func (s *skipWriter) WriteSkip(ctx context.Context, output store.IndexOutput, mwc *coreIndex.MultiLevelSkipListWriterContext) (int64, error) {
	skipOffset := output.GetFilePointer()

	w := utils.NewTextWriter(output)

	if err := w.Bytes(SKIP_LIST); err != nil {
		return 0, err
	}
	if err := w.NewLine(); err != nil {
		return 0, err
	}
	mwc.ResetSkip()
	return skipOffset, nil
}

func (s *skipWriter) WriteLevelLength(ctx context.Context, levelLength int64, output store.IndexOutput) error {
	if err := utils.WriteBytes(output, LEVEL_LENGTH); err != nil {
		return err
	}
	if err := utils.WriteString(output, fmt.Sprint(levelLength)); err != nil {
		return err
	}
	if err := utils.NewLine(output); err != nil {
		return err
	}
	return nil
}

func (s *skipWriter) WriteChildPointer(ctx context.Context, childPointer int64, skipBuffer store.DataOutput) error {
	if err := utils.WriteBytes(skipBuffer, CHILD_POINTER); err != nil {
		return err
	}
	if err := utils.WriteString(skipBuffer, fmt.Sprint(childPointer)); err != nil {
		return err
	}
	if err := utils.NewLine(skipBuffer); err != nil {
		return err
	}
	return nil
}

func (s *skipWriter) bufferSkip(doc int, docFilePointer int64, numDocs int, accumulator *coreIndex.CompetitiveImpactAccumulator) {
	s.curDoc = doc
	s.curDocFilePointer = docFilePointer
	s.curCompetitiveFreqNorms[0].AddAll(accumulator)
}

func (s *SkipWriter) ResetSkip() error {
	return s.sw.ResetSkip(s.mwc)
}

func (s *SkipWriter) WriteSkip(ctx context.Context, out store.IndexOutput) (int64, error) {
	return s.sw.WriteSkip(ctx, out, s.mwc)
}

func (s *SkipWriter) BufferSkip(ctx context.Context, doc int, docFilePointer int64, numDocs int,
	accumulator *coreIndex.CompetitiveImpactAccumulator) error {

	s.sw.bufferSkip(doc, docFilePointer, numDocs, accumulator)
	return s.mwc.BufferSkip(ctx, numDocs, s.sw)
}
