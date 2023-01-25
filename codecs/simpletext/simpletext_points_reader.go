package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.PointsReader = &SimpleTextPointsReader{}

type SimpleTextPointsReader struct {
	dataIn    store.IndexInput
	readState *index.SegmentReadState
	readers   map[string]*SimpleTextBKDReader
	scratch   *bytes.Buffer
}

func NewSimpleTextPointsReader(readState *index.SegmentReadState) (*SimpleTextPointsReader, error) {
	panic("")
}

func (s *SimpleTextPointsReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextPointsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextPointsReader) GetValues(field string) (index.PointValues, error) {
	//TODO implement me
	panic("implement me")
}
