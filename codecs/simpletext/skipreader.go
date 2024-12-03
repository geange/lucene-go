package simpletext

import (
	"bytes"
	"math"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

//var _ coreIndex.MultiLevelSkipListReader = &SkipReader{}

type SkipReader struct {
	mtx *coreIndex.MultiLevelSkipListReaderContext
	sr  *skipReader
}

func NewSkipReader(skipStream store.IndexInput) *SkipReader {
	mtx := coreIndex.NewMultiLevelSkipListReaderContext(
		skipStream, maxSkipLevels, BLOCK_SIZE, skipMultiplier)

	sr := newSkipReader(mtx)

	return &SkipReader{
		mtx: mtx,
		sr:  sr,
	}
}

func (s *SkipReader) SkipTo(target int) (int, error) {
	return s.sr.SkipTo(target, s.mtx)
}

func (s *SkipReader) GetNextSkipDoc() int {
	return s.sr.getNextSkipDoc(s.mtx)
}

func (s *SkipReader) GetNextSkipDocFP() int64 {
	return s.sr.getNextSkipDocFP()
}

func (s *SkipReader) GetImpacts() index.Impacts {
	return s.sr.getImpacts()
}

func (s *SkipReader) Reset(skipPointer int64, docFreq int) error {
	return s.sr.reset(skipPointer, docFreq, s.mtx)
}

// SkipReader
// This class reads skip lists with multiple levels.
// See TextFieldsWriter for the information about the encoding of the multi level skip lists.
type skipReader struct {
	//mtx *coreIndex.MultiLevelSkipListReaderContext

	scratchUTF16    *bytes.Buffer
	scratch         *bytes.Buffer
	impacts         index.Impacts
	perLevelImpacts [][]index.Impact
	nextSkipDocFP   int64
	numLevels       int
	hasSkipList     bool
}

func newSkipReader(mtx *coreIndex.MultiLevelSkipListReaderContext) *skipReader {
	reader := &skipReader{
		scratchUTF16:    new(bytes.Buffer),
		scratch:         new(bytes.Buffer),
		impacts:         nil,
		perLevelImpacts: make([][]index.Impact, 0),
		nextSkipDocFP:   -1,
		numLevels:       1,
		hasSkipList:     false,
	}

	reader.impacts = &innerImpacts{
		r:   reader,
		mtx: mtx,
	}

	reader.init(mtx)

	return reader
}

func (s *skipReader) reset(skipPointer int64, docFreq int, mtx *coreIndex.MultiLevelSkipListReaderContext) error {
	s.init(mtx)
	if skipPointer > 0 {
		if err := mtx.Init(skipPointer, docFreq, s); err != nil {
			return err
		}
		s.hasSkipList = true
	}
	return nil
}

func (s *skipReader) init(mtx *coreIndex.MultiLevelSkipListReaderContext) {
	s.nextSkipDocFP = -1
	s.numLevels = 1
	s.perLevelImpacts = make([][]index.Impact, mtx.MaxNumberOfSkipLevels())
	for i := range s.perLevelImpacts {
		impacts := make([]index.Impact, 0)
		impacts = append(impacts, coreIndex.NewImpact(math.MaxInt32, 1))
		s.perLevelImpacts[i] = impacts
	}
	s.hasSkipList = false
}

func (s *skipReader) SkipTo(target int, mtx *coreIndex.MultiLevelSkipListReaderContext) (int, error) {
	return mtx.SkipToWithSPI(target, s)
}

var _ index.Impacts = &innerImpacts{}

type innerImpacts struct {
	r   *skipReader
	mtx *coreIndex.MultiLevelSkipListReaderContext
}

func (i *innerImpacts) NumLevels() int {
	return i.r.numLevels
}

func (i *innerImpacts) GetDocIdUpTo(level int) int {
	return i.mtx.GetSkipDoc(level)
}

func (i *innerImpacts) GetImpacts(level int) []index.Impact {
	return i.r.perLevelImpacts[level]
}

var _ coreIndex.MultiLevelSkipListReaderSPI = &skipReader{}

func (s *skipReader) ReadSkipData(level int, skipStream store.IndexInput, mtx *coreIndex.MultiLevelSkipListReaderContext) (int64, error) {
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
			skipDoc = int(num) - mtx.GetSkipDoc(level)
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
	return int64(skipDoc), nil
}

func (s *skipReader) ReadLevelLength(skipStream store.IndexInput, mtx *coreIndex.MultiLevelSkipListReaderContext) (int64, error) {
	err := utils.ReadLine(skipStream, s.scratch)
	if err != nil {
		return 0, err
	}
	content := s.scratch.Bytes()
	return strconv.ParseInt(string(content[len(LEVEL_LENGTH):]), 10, 64)
}

func (s *skipReader) ReadChildPointer(skipStream store.IndexInput, mtx *coreIndex.MultiLevelSkipListReaderContext) (int64, error) {
	err := utils.ReadLine(skipStream, s.scratch)
	if err != nil {
		return 0, err
	}
	content := s.scratch.Bytes()
	return strconv.ParseInt(string(content[len(CHILD_POINTER):]), 10, 64)
}

func (s *skipReader) getNextSkipDoc(mtx *coreIndex.MultiLevelSkipListReaderContext) int {
	if !s.hasSkipList {
		return types.NO_MORE_DOCS
	}
	return mtx.GetSkipDoc(0)
}

func (s *skipReader) getNextSkipDocFP() int64 {
	return s.nextSkipDocFP
}

func (s *skipReader) getImpacts() index.Impacts {
	return s.impacts
}

func (s *skipReader) Reset(skipPointer int64, docFreq int, mtx *coreIndex.MultiLevelSkipListReaderContext) {
	s.init(mtx)
	if skipPointer > 0 {
		_ = mtx.Init(skipPointer, docFreq, s)
		s.hasSkipList = true
	}
}

func (s *skipReader) GetImpacts() index.Impacts {
	return s.impacts
}
