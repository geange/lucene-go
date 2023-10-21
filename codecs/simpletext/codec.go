package simpletext

import "github.com/geange/lucene-go/core/index"

func init() {
	index.RegisterCodec(&Codec{})
}

var _ index.Codec = &Codec{}

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

func (s *Codec) PostingsFormat() index.PostingsFormat {
	return s.postings
}

func (s *Codec) DocValuesFormat() index.DocValuesFormat {
	return s.dvFormat
}

func (s *Codec) StoredFieldsFormat() index.StoredFieldsFormat {
	return s.storedFields
}

func (s *Codec) TermVectorsFormat() index.TermVectorsFormat {
	return s.vectorsFormat
}

func (s *Codec) FieldInfosFormat() index.FieldInfosFormat {
	return s.fieldInfosFormat
}

func (s *Codec) SegmentInfoFormat() index.SegmentInfoFormat {
	return s.segmentInfos
}

func (s *Codec) NormsFormat() index.NormsFormat {
	return s.normsFormat
}

func (s *Codec) LiveDocsFormat() index.LiveDocsFormat {
	return s.liveDocs
}

func (s *Codec) CompoundFormat() index.CompoundFormat {
	return s.compoundFormat
}

func (s *Codec) PointsFormat() index.PointsFormat {
	return s.pointsFormat
}
