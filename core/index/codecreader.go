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

type BaseCodecReader struct {
	*BaseLeafReader

	CodecReaderDefaultSPI
}

func NewBaseCodecReader(reader CodecReader) *BaseCodecReader {
	codec := &BaseCodecReader{
		CodecReaderDefaultSPI: reader,
	}

	codec.BaseLeafReader = NewBaseLeafReader(reader)
	return codec
}

func (c *BaseCodecReader) DocumentV1(docID int, visitor document.StoredFieldVisitor) error {
	return c.GetFieldsReader().VisitDocument(docID, visitor)
}

func (c *BaseCodecReader) GetTermVectors(docID int) (Fields, error) {
	termVectorsReader := c.GetTermVectorsReader()
	if termVectorsReader == nil {
		return nil, nil
	}
	if err := c.checkBounds(docID); err != nil {
		return nil, err
	}
	return termVectorsReader.Get(docID)
}

func (c *BaseCodecReader) checkBounds(docID int) error {
	if docID < 0 || docID >= c.MaxDoc() {
		return errors.New("out-of-bounds for length")
	}
	return nil
}

func (c *BaseCodecReader) Terms(field string) (Terms, error) {
	return c.GetPostingsReader().Terms(field)
}

// returns the FieldInfo that corresponds to the given field and type, or
// null if the field does not exist, or not indexed as the requested
// DovDocValuesType.
func (c *BaseCodecReader) getDVField(field string, _type document.DocValuesType) *document.FieldInfo {
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

func (c *BaseCodecReader) GetNumericDocValues(field string) (NumericDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_NUMERIC)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetNumeric(fi)
}

func (c *BaseCodecReader) GetBinaryDocValues(field string) (BinaryDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_BINARY)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetBinary(fi)
}

func (c *BaseCodecReader) GetSortedDocValues(field string) (SortedDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSorted(fi)
}

func (c *BaseCodecReader) GetSortedNumericDocValues(field string) (SortedNumericDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED_NUMERIC)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSortedNumeric(fi)
}

func (c *BaseCodecReader) GetSortedSetDocValues(field string) (SortedSetDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSortedSet(fi)
}

func (c *BaseCodecReader) GetNormValues(field string) (NumericDocValues, error) {
	//ensureOpen();
	fi := c.GetFieldInfos().FieldInfo(field)
	if fi == nil || fi.HasNorms() == false {
		// Field does not exist or does not index norms
		return nil, nil
	}
	return c.GetNormsReader().GetNorms(fi)
}

func (c *BaseCodecReader) GetPointValues(field string) (types.PointValues, bool) {
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

func (c *BaseCodecReader) CheckIntegrity() error {
	return nil
}
