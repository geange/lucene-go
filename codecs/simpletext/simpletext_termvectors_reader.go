package simpletext

import "github.com/geange/lucene-go/core/index"

var _ index.TermVectorsReader = &SimpleTextTermVectorsReader{}

// SimpleTextTermVectorsReader Reads plain-text term vectors.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextTermVectorsReader struct {
}

func (s *SimpleTextTermVectorsReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermVectorsReader) Get(doc int) (index.Fields, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermVectorsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermVectorsReader) Clone() index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermVectorsReader) GetMergeInstance() index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
}
