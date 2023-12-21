package simpletext

import (
	"fmt"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
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

var _ index.MultiLevelSkipListWriter = &SkipWriter{}

type SkipWriter struct {
	*index.MultiLevelSkipListWriterDefault

	wroteHeaderPerLevelMap  map[int]bool
	curDoc                  int
	curDocFilePointer       int64
	curCompetitiveFreqNorms []*index.CompetitiveImpactAccumulator
}

func NewSkipWriter(writeState *index.SegmentWriteState) (*SkipWriter, error) {
	maxDoc, err := writeState.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	writer := &SkipWriter{
		wroteHeaderPerLevelMap:  make(map[int]bool),
		curDoc:                  0,
		curDocFilePointer:       0,
		curCompetitiveFreqNorms: make([]*index.CompetitiveImpactAccumulator, maxSkipLevels),
	}

	for i := range writer.curCompetitiveFreqNorms {
		writer.curCompetitiveFreqNorms[i] = index.NewCompetitiveImpactAccumulator()
	}

	writer.MultiLevelSkipListWriterDefault = index.NewMultiLevelSkipListWriterDefault(&index.MultiLevelSkipListWriterDefaultConfig{
		SkipInterval:      BLOCK_SIZE,
		SkipMultiplier:    skipMultiplier,
		MaxSkipLevels:     maxSkipLevels,
		DF:                maxDoc,
		WriteSkipData:     writer.WriteSkipData,
		WriteLevelLength:  writer.WriteLevelLength,
		WriteChildPointer: writer.WriteChildPointer,
	})
	writer.ResetSkip()
	return writer, nil
}

func (s *SkipWriter) WriteSkipData(level int, skipBuffer store.IndexOutput) error {
	wroteHeader := s.wroteHeaderPerLevelMap[level]

	w := utils.NewTextWriter(skipBuffer)

	if !wroteHeader {
		w.Bytes(LEVEL)
		w.String(fmt.Sprintf("%d", level))
		w.NewLine()

		s.wroteHeaderPerLevelMap[level] = true
	}
	w.Bytes(SKIP_DOC)
	w.String(fmt.Sprintf("%d", s.curDoc))
	w.NewLine()

	w.Bytes(SKIP_DOC_FP)
	w.String(fmt.Sprintf("%d", s.curDocFilePointer))
	w.NewLine()

	competitiveFreqNorms := s.curCompetitiveFreqNorms[level]
	impacts := competitiveFreqNorms.GetCompetitiveFreqNormPairs()
	//assert impacts.size() > 0;
	if level+1 < s.NumberOfSkipLevels {
		s.curCompetitiveFreqNorms[level+1].AddAll(competitiveFreqNorms)
	}
	w.Bytes(IMPACTS)
	w.NewLine()
	for _, impact := range impacts {
		w.Bytes(IMPACT)
		w.NewLine()

		w.Bytes(FREQ)
		w.String(fmt.Sprintf("%d", impact.Freq))
		w.NewLine()

		w.Bytes(NORM)
		w.String(fmt.Sprintf("%d", impact.Norm))
		w.NewLine()
	}
	w.Bytes(IMPACTS_END)
	w.NewLine()
	competitiveFreqNorms.Clear()

	return nil
}

func (s *SkipWriter) ResetSkip() {
	s.MultiLevelSkipListWriterDefault.ResetSkip()
	s.wroteHeaderPerLevelMap = map[int]bool{}
	s.curDoc = -1
	s.curDocFilePointer = -1
	for _, norm := range s.curCompetitiveFreqNorms {
		norm.Clear()
	}
}

func (s *SkipWriter) WriteSkip(output store.IndexOutput) (int64, error) {
	skipOffset := output.GetFilePointer()

	w := utils.NewTextWriter(output)

	if err := w.Bytes(SKIP_LIST); err != nil {
		return 0, err
	}
	if err := w.NewLine(); err != nil {
		return 0, err
	}
	if _, err := s.MultiLevelSkipListWriterDefault.WriteSkip(output); err != nil {
		return 0, err
	}
	return skipOffset, nil
}

//func (s *SimpleTextSkipWriter) BufferSkipV1(doc int, docFilePointer int64, numDocs int, competitiveImpactAccumulator *index.CompetitiveImpactAccumulator) error {
//	//assert doc > curDoc;
//	s.curDoc = doc
//	s.curDocFilePointer = docFilePointer
//	s.curCompetitiveFreqNorms[0].AddAll(competitiveImpactAccumulator)
//	return s.BufferSkip(numDocs)
//}

func (s *SkipWriter) WriteLevelLength(levelLength int64, output store.IndexOutput) error {
	utils.WriteBytes(output, LEVEL_LENGTH)
	utils.WriteString(output, fmt.Sprintf("%d", levelLength))
	utils.NewLine(output)
	return nil
}

func (s *SkipWriter) WriteChildPointer(childPointer int64, skipBuffer store.DataOutput) error {
	utils.WriteBytes(skipBuffer, CHILD_POINTER)
	utils.WriteString(skipBuffer, fmt.Sprintf("%d", childPointer))
	utils.NewLine(skipBuffer)
	return nil
}

func (s *SkipWriter) bufferSkip(doc int, docFilePointer int64, numDocs int, accumulator *index.CompetitiveImpactAccumulator) error {
	s.curDoc = doc
	s.curDocFilePointer = docFilePointer
	s.curCompetitiveFreqNorms[0].AddAll(accumulator)
	return s.BufferSkip(numDocs)
}
