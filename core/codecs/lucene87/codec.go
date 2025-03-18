package lucene87

import (
	"github.com/geange/lucene-go/core/codecs/lucene50"
	"github.com/geange/lucene-go/core/codecs/lucene60"
	"github.com/geange/lucene-go/core/codecs/lucene80"
	"github.com/geange/lucene-go/core/codecs/lucene84"
	"github.com/geange/lucene-go/core/codecs/lucene86"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.Codec = &Codec{}

type Codec struct {
	vectorsFormat      index.TermVectorsFormat
	fieldInfosFormat   index.FieldInfosFormat
	segmentInfosFormat index.SegmentInfoFormat
	liveDocsFormat     index.LiveDocsFormat
	compoundFormat     index.CompoundFormat
	pointsFormat       index.PointsFormat
	defaultFormat      index.PostingsFormat
	postingsFormat     index.PostingsFormat
	docValuesFormat    index.DocValuesFormat
	storedFieldsFormat index.StoredFieldsFormat
	defaultDVFormat    index.DocValuesFormat
}

func NewCodec(mode Mode) *Codec {
	return &Codec{
		vectorsFormat:      lucene50.NewTermVectorsFormat(),
		fieldInfosFormat:   lucene60.NewFieldInfosFormat(),
		segmentInfosFormat: lucene86.NewSegmentInfoFormat(),
		liveDocsFormat:     lucene50.NewLiveDocsFormat(),
		compoundFormat:     lucene50.NewCompoundFormat(),
		pointsFormat:       lucene86.NewPointsFormat(),
		defaultFormat:      lucene84.NewPostingsFormat(),
		defaultDVFormat:    lucene80.NewDocValuesFormat(),
		postingsFormat:     lucene84.NewPostingsFormat(),
		docValuesFormat:    lucene80.NewDocValuesFormat(),
		storedFieldsFormat: NewStoredFieldsFormat(mode),
	}

	//res.postingsFormat = res.defaultFormat
	//res.docValuesFormat = res.defaultDVFormat
}

func (c *Codec) GetName() string {
	return "Lucene87Codec"
}

func (c *Codec) PostingsFormat() index.PostingsFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) DocValuesFormat() index.DocValuesFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) StoredFieldsFormat() index.StoredFieldsFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) TermVectorsFormat() index.TermVectorsFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) FieldInfosFormat() index.FieldInfosFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) SegmentInfoFormat() index.SegmentInfoFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) NormsFormat() index.NormsFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) LiveDocsFormat() index.LiveDocsFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) CompoundFormat() index.CompoundFormat {
	//TODO implement me
	panic("implement me")
}

func (c *Codec) PointsFormat() index.PointsFormat {
	//TODO implement me
	panic("implement me")
}
