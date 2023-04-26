package simpletext

import "github.com/geange/lucene-go/core/index"

func init() {
	index.RegisterCodec(&SimpleTextCodec{})
}

var _ index.Codec = &SimpleTextCodec{}

// SimpleTextCodec plain text index format.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextCodec struct {
	postings         *SimpleTextPostingsFormat
	storedFields     *SimpleTextStoredFieldsFormat
	segmentInfos     *SimpleTextSegmentInfoFormat
	fieldInfosFormat *SimpleTextFieldInfosFormat
	vectorsFormat    *SimpleTextTermVectorsFormat
	normsFormat      *SimpleTextNormsFormat
	liveDocs         *SimpleTextLiveDocsFormat
	dvFormat         *SimpleTextDocValuesFormat
	compoundFormat   *SimpleTextCompoundFormat
	pointsFormat     *SimpleTextPointsFormat
}

func NewSimpleTextCodec() *SimpleTextCodec {
	return &SimpleTextCodec{
		postings:         NewSimpleTextPostingsFormat(),
		storedFields:     NewSimpleTextStoredFieldsFormat(),
		segmentInfos:     NewSimpleTextSegmentInfoFormat(),
		fieldInfosFormat: NewSimpleTextFieldInfosFormat(),
		vectorsFormat:    NewSimpleTextTermVectorsFormat(),
		normsFormat:      NewSimpleTextNormsFormat(),
		liveDocs:         NewSimpleTextLiveDocsFormat(),
		dvFormat:         NewSimpleTextDocValuesFormat(),
		compoundFormat:   NewSimpleTextCompoundFormat(),
		pointsFormat:     NewSimpleTextPointsFormat(),
	}
}

func (s *SimpleTextCodec) GetName() string {
	return "SimpleText"
}

func (s *SimpleTextCodec) PostingsFormat() index.PostingsFormat {
	return s.postings
}

func (s *SimpleTextCodec) DocValuesFormat() index.DocValuesFormat {
	return s.dvFormat
}

func (s *SimpleTextCodec) StoredFieldsFormat() index.StoredFieldsFormat {
	return s.storedFields
}

func (s *SimpleTextCodec) TermVectorsFormat() index.TermVectorsFormat {
	return s.vectorsFormat
}

func (s *SimpleTextCodec) FieldInfosFormat() index.FieldInfosFormat {
	return s.fieldInfosFormat
}

func (s *SimpleTextCodec) SegmentInfoFormat() index.SegmentInfoFormat {
	return s.segmentInfos
}

func (s *SimpleTextCodec) NormsFormat() index.NormsFormat {
	return s.normsFormat
}

func (s *SimpleTextCodec) LiveDocsFormat() index.LiveDocsFormat {
	return s.liveDocs
}

func (s *SimpleTextCodec) CompoundFormat() index.CompoundFormat {
	return s.compoundFormat
}

func (s *SimpleTextCodec) PointsFormat() index.PointsFormat {
	return s.pointsFormat
}
