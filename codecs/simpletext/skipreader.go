package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/codecs/utils"
	"math"
	"strconv"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ coreIndex.MultiLevelSkipListReader = &SkipReader{}

// SkipReader
// This class reads skip lists with multiple levels.
// See TextFieldsWriter for the information about the encoding of the multi level skip lists.
type SkipReader struct {
	*coreIndex.BaseMultiLevelSkipListReader

	scratchUTF16    *bytes.Buffer
	scratch         *bytes.Buffer
	impacts         index.Impacts
	perLevelImpacts [][]index.Impact
	nextSkipDocFP   int64
	numLevels       int
	hasSkipList     bool
}

func NewSkipReader(skipStream store.IndexInput) *SkipReader {
	mReader := coreIndex.NewBaseMultiLevelSkipListReader(
		skipStream, maxSkipLevels, BLOCK_SIZE, skipMultiplier)

	reader := &SkipReader{
		BaseMultiLevelSkipListReader: mReader,
		scratchUTF16:                 new(bytes.Buffer),
		scratch:                      new(bytes.Buffer),
		impacts:                      nil,
		perLevelImpacts:              make([][]index.Impact, 0),
		nextSkipDocFP:                -1,
		numLevels:                    1,
		hasSkipList:                  false,
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
	s.perLevelImpacts = make([][]index.Impact, s.MaxNumberOfSkipLevels())
	for i := range s.perLevelImpacts {
		impacts := make([]index.Impact, 0)
		impacts = append(impacts, coreIndex.NewImpact(math.MaxInt32, 1))
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
	return i.r.GetSkipDoc(level)
}

func (i *innerImpacts) GetImpacts(level int) []index.Impact {
	return i.r.perLevelImpacts[level]
}

func (s *SkipReader) ReadSkipData(level int, skipStream store.IndexInput) (int, error) {
	s.perLevelImpacts[level] = nil

	skipDoc := types.NO_MORE_DOCS

	input := store.NewBufferedChecksumIndexInput(skipStream)
	freq := int64(1)

	for {
		err := utils.ReadLine(input, s.scratch)
		if err != nil {
			return 0, err
		}

		content := s.scratch.Bytes()

		if bytes.Equal(content, FIELDS_END) {
			err := utils.CheckFooter(input)
			if err != nil {
				return 0, err
			}
			break
		}

		if bytes.Equal(content, IMPACTS_END) ||
			bytes.Equal(content, FIELDS_TERM) ||
			bytes.Equal(content, FIELDS_FIELD) {
			break
		}

		if bytes.Equal(content, SKIP_LIST) {
			continue
		}

		if bytes.Equal(content, SKIP_DOC) {
			offset := len(SKIP_DOC)
			num, err := strconv.ParseInt(string(content[offset:]), 10, 64)
			if err != nil {
				return 0, err
			}
			//Because the MultiLevelSkipListReader stores doc id delta,but simple text codec stores doc id
			skipDoc = int(num) - (s.GetSkipDoc(level))
			continue
		}

		if bytes.Equal(content, SKIP_DOC_FP) {
			offset := len(SKIP_DOC_FP)
			num, err := strconv.ParseInt(string(content[offset:]), 10, 64)
			if err != nil {
				return 0, err
			}
			s.nextSkipDocFP = num
			continue
		}

		if bytes.Equal(content, IMPACTS) {
			continue
		}

		if bytes.Equal(content, FREQ) {
			offset := len(FREQ)
			num, err := strconv.ParseInt(string(content[offset:]), 10, 64)
			if err != nil {
				return 0, err
			}
			freq = num
			continue
		}

		if bytes.Equal(content, NORM) {
			offset := len(NORM)
			norm, err := strconv.ParseInt(string(content[offset:]), 10, 64)
			if err != nil {
				return 0, err
			}
			impact := coreIndex.NewImpact(int(freq), norm)
			s.perLevelImpacts[level] = append(s.perLevelImpacts[level], impact)
			continue
		}
	}
	return skipDoc, nil
}

func (s *SkipReader) ReadLevelLength(skipStream store.IndexInput) (int64, error) {
	err := utils.ReadLine(skipStream, s.scratch)
	if err != nil {
		return 0, err
	}
	content := s.scratch.Bytes()
	return strconv.ParseInt(string(content[len(LEVEL_LENGTH):]), 10, 64)
}

func (s *SkipReader) ReadChildPointer(skipStream store.IndexInput) (int64, error) {
	err := utils.ReadLine(skipStream, s.scratch)
	if err != nil {
		return 0, err
	}
	content := s.scratch.Bytes()
	return strconv.ParseInt(string(content[len(CHILD_POINTER):]), 10, 64)
}

func (s *SkipReader) getNextSkipDoc() int {
	if !s.hasSkipList {
		return types.NO_MORE_DOCS
	}
	return s.GetSkipDoc(0)
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
		s.BaseMultiLevelSkipListReader.Init(skipPointer, docFreq)
		s.hasSkipList = true
	}
}

func (s *SkipReader) GetImpacts() index.Impacts {
	return s.impacts
}
