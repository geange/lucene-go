package index

import (
	"errors"
	"github.com/geange/lucene-go/core/types"

	"github.com/geange/lucene-go/core/document"
)

type CodecReader interface {
	LeafReader

	// GetFieldsReader Expert: retrieve thread-private StoredFieldsReader
	// lucene.internal
	GetFieldsReader() StoredFieldsReader

	// GetTermVectorsReader Expert: retrieve thread-private TermVectorsReader
	// lucene.internal
	GetTermVectorsReader() TermVectorsReader

	// GetNormsReader Expert: retrieve underlying NormsProducer
	// lucene.internal
	GetNormsReader() NormsProducer

	// GetDocValuesReader Expert: retrieve underlying DocValuesProducer
	// lucene.internal
	GetDocValuesReader() DocValuesProducer

	// GetPostingsReader Expert: retrieve underlying FieldsProducer
	// lucene.internal
	GetPostingsReader() FieldsProducer

	// GetPointsReader Expert: retrieve underlying PointsReader
	// lucene.internal
	GetPointsReader() PointsReader
}

type CodecReaderDefaultSPI interface {
	GetFieldsReader() StoredFieldsReader
	GetTermVectorsReader() TermVectorsReader
	GetPostingsReader() FieldsProducer
	GetFieldInfos() *FieldInfos
	MaxDoc() int
	GetDocValuesReader() DocValuesProducer
	GetNormsReader() NormsProducer
	GetPointsReader() PointsReader
}

type CodecReaderDefault struct {
	*LeafReaderBase

	CodecReaderDefaultSPI
}

func NewCodecReaderDefault(reader CodecReader) *CodecReaderDefault {
	codec := &CodecReaderDefault{
		CodecReaderDefaultSPI: reader,
	}

	codec.LeafReaderBase = NewLeafReaderDefault(reader)
	return codec
}

func (c *CodecReaderDefault) DocumentV1(docID int, visitor document.StoredFieldVisitor) error {
	return c.GetFieldsReader().VisitDocument(docID, visitor)
}

func (c *CodecReaderDefault) GetTermVectors(docID int) (Fields, error) {
	termVectorsReader := c.GetTermVectorsReader()
	if termVectorsReader == nil {
		return nil, nil
	}
	if err := c.checkBounds(docID); err != nil {
		return nil, err
	}
	return termVectorsReader.Get(docID)
}

func (c *CodecReaderDefault) checkBounds(docID int) error {
	if docID < 0 || docID >= c.MaxDoc() {
		return errors.New("out-of-bounds for length")
	}
	return nil
}

func (c *CodecReaderDefault) Terms(field string) (Terms, error) {
	return c.GetPostingsReader().Terms(field)
}

// returns the FieldInfo that corresponds to the given field and type, or
// null if the field does not exist, or not indexed as the requested
// DovDocValuesType.
func (c *CodecReaderDefault) getDVField(field string, _type document.DocValuesType) *document.FieldInfo {
	fi := c.GetFieldInfos().FieldInfo(field)
	if fi == nil {
		// Field does not exist
		return nil
	}
	if fi.GetDocValuesType() == document.DOC_VALUES_TYPE_NONE {
		// Field was not indexed with doc values
		return nil
	}
	if fi.GetDocValuesType() != _type {
		// Field DocValues are different than requested type
		return nil
	}
	return fi
}

func (c *CodecReaderDefault) GetNumericDocValues(field string) (NumericDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_NUMERIC)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetNumeric(fi)
}

func (c *CodecReaderDefault) GetBinaryDocValues(field string) (BinaryDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_BINARY)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetBinary(fi)
}

func (c *CodecReaderDefault) GetSortedDocValues(field string) (SortedDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSorted(fi)
}

func (c *CodecReaderDefault) GetSortedNumericDocValues(field string) (SortedNumericDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED_NUMERIC)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSortedNumeric(fi)
}

func (c *CodecReaderDefault) GetSortedSetDocValues(field string) (SortedSetDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSortedSet(fi)
}

func (c *CodecReaderDefault) GetNormValues(field string) (NumericDocValues, error) {
	//ensureOpen();
	fi := c.GetFieldInfos().FieldInfo(field)
	if fi == nil || fi.HasNorms() == false {
		// Field does not exist or does not index norms
		return nil, nil
	}
	return c.GetNormsReader().GetNorms(fi)
}

func (c *CodecReaderDefault) GetPointValues(field string) (types.PointValues, bool) {
	//ensureOpen();
	fi := c.GetFieldInfos().FieldInfo(field)
	if fi == nil || fi.GetPointDimensionCount() == 0 {
		// Field does not exist or does not index points
		return nil, false
	}
	values, err := c.GetPointsReader().GetValues(field)
	if err != nil {
		return nil, false
	}
	return values, true
}

func (c *CodecReaderDefault) CheckIntegrity() error {
	return nil
}
