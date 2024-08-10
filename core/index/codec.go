package index

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var codesPool = make(map[string]index.Codec)

func RegisterCodec(codec index.Codec) {
	codesPool[codec.GetName()] = codec
}

func GetCodecByName(name string) (index.Codec, bool) {
	codec, exist := codesPool[name]
	return codec, exist
}

type BaseCompoundDirectory struct {
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation exception")
)

func (*BaseCompoundDirectory) DeleteFile(ctx context.Context, name string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) Rename(ctx context.Context, source, dest string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) SyncMetaData(ctx context.Context) error {
	return nil
}

func (*BaseCompoundDirectory) CreateOutput(ctx context.Context, name string) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) CreateTempOutput(ctx context.Context, prefix, suffix string) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) Sync(names map[string]struct{}) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) ObtainLock(name string) (store.Lock, error) {
	return nil, ErrUnsupportedOperation
}

type BaseCodecReader struct {
	*BaseLeafReader

	index.CodecReaderSPI
}

func NewBaseCodecReader(reader index.CodecReader) *BaseCodecReader {
	codec := &BaseCodecReader{
		CodecReaderSPI: reader,
	}

	codec.BaseLeafReader = NewBaseLeafReader(reader)
	return codec
}

func (c *BaseCodecReader) DocumentWithVisitor(docID int, visitor document.StoredFieldVisitor) error {
	return c.GetFieldsReader().VisitDocument(nil, docID, visitor)
}

func (c *BaseCodecReader) GetTermVectors(docID int) (index.Fields, error) {
	termVectorsReader := c.GetTermVectorsReader()
	if termVectorsReader == nil {
		return nil, nil
	}
	if err := c.checkBounds(docID); err != nil {
		return nil, err
	}
	return termVectorsReader.Get(nil, docID)
}

func (c *BaseCodecReader) checkBounds(docID int) error {
	if docID < 0 || docID >= c.MaxDoc() {
		return errors.New("out-of-bounds for length")
	}
	return nil
}

func (c *BaseCodecReader) Terms(field string) (index.Terms, error) {
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

func (c *BaseCodecReader) GetNumericDocValues(field string) (index.NumericDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_NUMERIC)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetNumeric(nil, fi)
}

func (c *BaseCodecReader) GetBinaryDocValues(field string) (index.BinaryDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_BINARY)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetBinary(nil, fi)
}

func (c *BaseCodecReader) GetSortedDocValues(field string) (index.SortedDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSorted(nil, fi)
}

func (c *BaseCodecReader) GetSortedNumericDocValues(field string) (index.SortedNumericDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED_NUMERIC)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSortedNumeric(nil, fi)
}

func (c *BaseCodecReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	//ensureOpen();
	fi := c.getDVField(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if fi == nil {
		return nil, nil
	}
	return c.GetDocValuesReader().GetSortedSet(nil, fi)
}

func (c *BaseCodecReader) GetNormValues(field string) (index.NumericDocValues, error) {
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
	values, err := c.GetPointsReader().GetValues(nil, field)
	if err != nil {
		return nil, false
	}
	return values, true
}

func (c *BaseCodecReader) CheckIntegrity() error {
	return nil
}
