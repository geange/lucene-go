package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/types"
	"math"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.MultiLevelSkipListReader = &SkipReader{}

// SkipReader This class reads skip lists with multiple levels.
// See TextFieldsWriter for the information about the encoding of the multi level skip lists.
type SkipReader struct {
	*index.MultiLevelSkipListReaderDefault

	scratchUTF16    *bytes.Buffer
	scratch         *bytes.Buffer
	impacts         index.Impacts
	perLevelImpacts [][]*index.Impact
	nextSkipDocFP   int64
	numLevels       int
	hasSkipList     bool
}

func NewSkipReader(skipStream store.IndexInput) *SkipReader {
	mReader := index.NewMultiLevelSkipListReaderDefault(
		skipStream, maxSkipLevels, BLOCK_SIZE, skipMultiplier)

	reader := &SkipReader{
		MultiLevelSkipListReaderDefault: mReader,
		scratchUTF16:                    new(bytes.Buffer),
		scratch:                         new(bytes.Buffer),
		impacts:                         nil,
		perLevelImpacts:                 make([][]*index.Impact, 0),
		nextSkipDocFP:                   -1,
		numLevels:                       1,
		hasSkipList:                     false,
	}

	reader.impacts = &innerImpacts{reader}

	reader.init()

	return reader
}

func (s *SkipReader) reset(skipPointer int64, docFreq int) error {
	s.init()
	if skipPointer > 0 {
		if err := s.Init(skipPointer, docFreq); err != nil {
			return err
		}
		s.hasSkipList = true
	}
	return nil
}

func (s *SkipReader) init() {
	s.nextSkipDocFP = -1
	s.numLevels = 1
	s.perLevelImpacts = make([][]*index.Impact, s.MaxNumberOfSkipLevels)
	for i := range s.perLevelImpacts {
		impacts := make([]*index.Impact, 0)
		impacts = append(impacts, index.NewImpact(math.MaxInt32, 1))
		s.perLevelImpacts[i] = impacts
	}
	s.hasSkipList = false
}

var _ index.Impacts = &innerImpacts{}

type innerImpacts struct {
	r *SkipReader
}

func (i *innerImpacts) NumLevels() int {
	return i.r.numLevels
}

func (i *innerImpacts) GetDocIdUpTo(level int) int {
	return i.r.SkipDoc[level]
}

func (i *innerImpacts) GetImpacts(level int) []*index.Impact {
	return i.r.perLevelImpacts[level]
}

func (s *SkipReader) ReadSkipData(level int, skipStream store.IndexInput) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SkipReader) getNextSkipDoc() int {
	if !s.hasSkipList {
		return types.NO_MORE_DOCS
	}
	return s.SkipDoc[0]
}

func (s *SkipReader) getNextSkipDocFP() int64 {
	return s.nextSkipDocFP
}

func (s *SkipReader) getImpacts() index.Impacts {
	return s.impacts
}

func (s *SkipReader) Reset(skipPointer int64, docFreq int) {
	s.init()
	if skipPointer > 0 {
		s.MultiLevelSkipListReaderDefault.Init(skipPointer, docFreq)
		s.hasSkipList = true
	}
}

func (s *SkipReader) GetImpacts() index.Impacts {
	return s.impacts
}
