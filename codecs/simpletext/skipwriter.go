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

// TODO: fix
//var _ index.MultiLevelSkipListWriter = &SkipWriter{}

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
		// TODO: remove
		fmt.Println(maxDoc)
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

	// TODO: fix
	/*
		writer.MultiLevelSkipListWriterDefault = index.NewMultiLevelSkipListWriterDefault(&index.MultiLevelSkipListWriterDefaultConfig{
			SkipInterval:      BLOCK_SIZE,
			SkipMultiplier:    skipMultiplier,
			MaxSkipLevels:     maxSkipLevels,
			DF:                maxDoc,
			WriteSkipData:     writer.WriteSkipData,
			WriteLevelLength:  writer.WriteLevelLength,
			WriteChildPointer: writer.WriteChildPointer,
		})

	*/
	writer.ResetSkip()
	return writer, nil
}

func (s *SkipWriter) WriteSkipData(level int, skipBuffer store.IndexOutput) error {
	wroteHeader := s.wroteHeaderPerLevelMap[level]

	w := utils.NewTextWriter(skipBuffer)

	if !wroteHeader {
		if err := w.Bytes(LEVEL); err != nil {
			return err
		}
		if err := w.String(fmt.Sprintf("%d", level)); err != nil {
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
	if err := w.String(fmt.Sprintf("%d", s.curDoc)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	if err := w.Bytes(SKIP_DOC_FP); err != nil {
		return err
	}
	if err := w.String(fmt.Sprintf("%d", s.curDocFilePointer)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	competitiveFreqNorms := s.curCompetitiveFreqNorms[level]
	impacts := competitiveFreqNorms.GetCompetitiveFreqNormPairs()
	//assert impacts.size() > 0;
	if level+1 < s.NumberOfSkipLevels {
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
		if err := w.String(fmt.Sprintf("%d", impact.Freq)); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		if err := w.Bytes(NORM); err != nil {
			return err
		}
		if err := w.String(fmt.Sprintf("%d", impact.Norm)); err != nil {
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

func (s *SkipWriter) ResetSkip() {
	// TODO: fix
	//s.MultiLevelSkipListWriterDefault.ResetSkip()
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

	// TODO: fix
	/*
		if _, err := s.MultiLevelSkipListWriterDefault.WriteSkip(output); err != nil {
			return 0, err
		}

	*/
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
	if err := utils.WriteBytes(output, LEVEL_LENGTH); err != nil {
		return err
	}
	if err := utils.WriteString(output, fmt.Sprintf("%d", levelLength)); err != nil {
		return err
	}
	if err := utils.NewLine(output); err != nil {
		return err
	}
	return nil
}

func (s *SkipWriter) WriteChildPointer(childPointer int64, skipBuffer store.DataOutput) error {
	if err := utils.WriteBytes(skipBuffer, CHILD_POINTER); err != nil {
		return err
	}
	if err := utils.WriteString(skipBuffer, fmt.Sprintf("%d", childPointer)); err != nil {
		return err
	}
	if err := utils.NewLine(skipBuffer); err != nil {
		return err
	}
	return nil
}

func (s *SkipWriter) bufferSkip(doc int, docFilePointer int64, numDocs int, accumulator *index.CompetitiveImpactAccumulator) error {
	s.curDoc = doc
	s.curDocFilePointer = docFilePointer
	s.curCompetitiveFreqNorms[0].AddAll(accumulator)
	return nil
	// TODO: fix
	/*
		return s.BufferSkip(numDocs)

	*/
}
