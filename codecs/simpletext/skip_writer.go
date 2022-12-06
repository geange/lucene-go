package simpletext

import (
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

const (
	skipMultiplier = 3
	maxSkipLevels  = 4

	BLOCK_SIZE = 8
)

var (
	SimpleTextSkipWriterToken = struct {
		SKIP_LIST     []byte
		LEVEL_LENGTH  []byte
		LEVEL         []byte
		SKIP_DOC      []byte
		SKIP_DOC_FP   []byte
		IMPACTS       []byte
		IMPACT        []byte
		FREQ          []byte
		NORM          []byte
		IMPACTS_END   []byte
		CHILD_POINTER []byte
	}{
		SKIP_LIST:     []byte("    skipList "),
		LEVEL_LENGTH:  []byte("      levelLength "),
		LEVEL:         []byte("      level "),
		SKIP_DOC:      []byte("        skipDoc "),
		SKIP_DOC_FP:   []byte("        skipDocFP "),
		IMPACTS:       []byte("        impacts "),
		IMPACT:        []byte("          impact "),
		FREQ:          []byte("            freq "),
		NORM:          []byte("            norm "),
		IMPACTS_END:   []byte("        impactsEnd "),
		CHILD_POINTER: []byte("        childPointer "),
	}

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

var _ index.MultiLevelSkipListWriter = &SimpleTextSkipWriter{}

type SimpleTextSkipWriter struct {
	wroteHeaderPerLevelMap  *hashmap.Map
	curDoc                  int
	curDocFilePointer       int64
	curCompetitiveFreqNorms []index.CompetitiveImpactAccumulator
}

func (s *SimpleTextSkipWriter) WriteSkipData(level int, skipBuffer store.IndexOutput) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipWriter) Init() {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipWriter) ResetSkip() error {
	s.wroteHeaderPerLevelMap.Clear()
	s.curDoc = -1
	s.curDocFilePointer = -1
	for _, norm := range s.curCompetitiveFreqNorms {
		norm.Clear()
	}
	return nil
}

func (s *SimpleTextSkipWriter) BufferSkip(df int) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipWriter) WriteSkip(output store.IndexOutput) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipWriter) WriteLevelLength(levelLength int64, output store.IndexOutput) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipWriter) WriteChildPointer(childPointer int64, skipBuffer store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipWriter) bufferSkip(doc int, docFilePointer int64, numDocs int, accumulator *index.CompetitiveImpactAccumulator) error {
	s.curDoc = doc
	s.curDocFilePointer = docFilePointer
	s.curCompetitiveFreqNorms[0].AddAll(accumulator)
	return s.BufferSkip(numDocs)
}
