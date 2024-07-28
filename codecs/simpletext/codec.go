package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	index2 "github.com/geange/lucene-go/core/interface/index"
)

func init() {
	index.RegisterCodec(&Codec{})
}

var _ index2.Codec = &Codec{}

// Codec plain text index format.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type Codec struct {
	postings         *PostingsFormat
	storedFields     *StoredFieldsFormat
	segmentInfos     *SegmentInfoFormat
	fieldInfosFormat *FieldInfosFormat
	vectorsFormat    *TermVectorsFormat
	normsFormat      *NormsFormat
	liveDocs         *LiveDocsFormat
	dvFormat         *DocValuesFormat
	compoundFormat   *CompoundFormat
	pointsFormat     *PointsFormat
}

func NewCodec() *Codec {
	return &Codec{
		postings:         NewPostingsFormat(),
		storedFields:     NewStoredFieldsFormat(),
		segmentInfos:     NewSegmentInfoFormat(),
		fieldInfosFormat: NewSimpleTextFieldInfosFormat(),
		vectorsFormat:    NewTermVectorsFormat(),
		normsFormat:      NewNormsFormat(),
		liveDocs:         NewLiveDocsFormat(),
		dvFormat:         NewSimpleTextDocValuesFormat(),
		compoundFormat:   NewCompoundFormat(),
		pointsFormat:     NewPointsFormat(),
	}
}

func (s *Codec) GetName() string {
	return "SimpleText"
}

func (s *Codec) PostingsFormat() index2.PostingsFormat {
	return s.postings
}

func (s *Codec) DocValuesFormat() index2.DocValuesFormat {
	return s.dvFormat
}

func (s *Codec) StoredFieldsFormat() index2.StoredFieldsFormat {
	return s.storedFields
}

func (s *Codec) TermVectorsFormat() index2.TermVectorsFormat {
	return s.vectorsFormat
}

func (s *Codec) FieldInfosFormat() index2.FieldInfosFormat {
	return s.fieldInfosFormat
}

func (s *Codec) SegmentInfoFormat() index2.SegmentInfoFormat {
	return s.segmentInfos
}

func (s *Codec) NormsFormat() index2.NormsFormat {
	return s.normsFormat
}

func (s *Codec) LiveDocsFormat() index2.LiveDocsFormat {
	return s.liveDocs
}

func (s *Codec) CompoundFormat() index2.CompoundFormat {
	return s.compoundFormat
}

func (s *Codec) PointsFormat() index2.PointsFormat {
	return s.pointsFormat
}
