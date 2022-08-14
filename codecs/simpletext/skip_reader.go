package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

var _ index.MultiLevelSkipListReader = &SimpleTextSkipReader{}

// SimpleTextSkipReader This class reads skip lists with multiple levels.
// See SimpleTextFieldsWriter for the information about the encoding of the multi level skip lists.
type SimpleTextSkipReader struct {
	index.MultiLevelSkipListReaderImp

	scratchUTF16    *util.CharsRefBuilder
	scratch         *util.BytesRefBuilder
	impacts         index.Impact
	perLevelImpacts [][]*index.Impact
	nextSkipDocFP   int64
	numLevels       int
	hasSkipList     bool
}

func (s *SimpleTextSkipReader) ReadSkipData(level int, skipStream store.IndexInput) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextSkipReader) getNextSkipDoc() int {
	if !s.hasSkipList {
		return index.NO_MORE_DOCS
	}
	return s.SkipDoc[0]
}

func (s *SimpleTextSkipReader) getNextSkipDocFP() int64 {
	return s.nextSkipDocFP
}

func (s *SimpleTextSkipReader) getImpacts() index.Impacts {
	return s.impacts
}
