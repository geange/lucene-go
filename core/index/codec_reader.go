package index

import "github.com/geange/lucene-go/core/document"

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

type CodecReaderDefault struct {
	getFieldsReader func() StoredFieldsReader
	maxDoc          func() int
}

func (c *CodecReaderDefault) DocumentV1(docID int, visitor document.StoredFieldVisitor) error {
	return c.getFieldsReader().VisitDocument(docID, visitor)
}
