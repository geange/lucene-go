package lucene87

import (
	"github.com/geange/lucene-go/core/codecs/lucene50"
	"github.com/geange/lucene-go/core/codecs/lucene60"
	"github.com/geange/lucene-go/core/codecs/lucene80"
	"github.com/geange/lucene-go/core/codecs/lucene84"
	"github.com/geange/lucene-go/core/codecs/lucene86"
	"github.com/geange/lucene-go/core/codecs/perfield"
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
	normsFormat        index.NormsFormat
}

func NewCodec(mode Mode) *Codec {
	codec := &Codec{
		vectorsFormat:      lucene50.NewTermVectorsFormat(),
		fieldInfosFormat:   lucene60.NewFieldInfosFormat(),
		segmentInfosFormat: lucene86.NewSegmentInfoFormat(),
		liveDocsFormat:     lucene50.NewLiveDocsFormat(),
		compoundFormat:     lucene50.NewCompoundFormat(),
		pointsFormat:       lucene86.NewPointsFormat(),
		defaultFormat:      lucene84.NewPostingsFormat(),
		defaultDVFormat:    lucene80.NewDocValuesFormat(),
		postingsFormat:     nil,
		docValuesFormat:    lucene80.NewDocValuesFormat(),
		storedFieldsFormat: NewStoredFieldsFormat(mode),
		normsFormat:        lucene80.NewNormsFormat(),
	}

	codec.postingsFormat = perfield.NewPostingsFormat(func(field string) index.PostingsFormat {
		return codec.defaultFormat
	})

	return codec
}

func (c *Codec) GetName() string {
	return "Lucene87Codec"
}

func (c *Codec) PostingsFormat() index.PostingsFormat {
	return c.postingsFormat
}

func (c *Codec) DocValuesFormat() index.DocValuesFormat {
	return c.docValuesFormat
}

func (c *Codec) StoredFieldsFormat() index.StoredFieldsFormat {
	return c.storedFieldsFormat
}

func (c *Codec) TermVectorsFormat() index.TermVectorsFormat {
	return c.vectorsFormat
}

func (c *Codec) FieldInfosFormat() index.FieldInfosFormat {
	return c.fieldInfosFormat
}

func (c *Codec) SegmentInfoFormat() index.SegmentInfoFormat {
	return c.segmentInfosFormat
}

func (c *Codec) NormsFormat() index.NormsFormat {
	return c.normsFormat
}

func (c *Codec) LiveDocsFormat() index.LiveDocsFormat {
	return c.liveDocsFormat
}

func (c *Codec) CompoundFormat() index.CompoundFormat {
	return c.compoundFormat
}

func (c *Codec) PointsFormat() index.PointsFormat {
	return c.pointsFormat
}
