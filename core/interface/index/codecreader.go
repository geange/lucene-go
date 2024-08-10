package index

type CodecReaderSPI interface {
	GetFieldsReader() StoredFieldsReader
	GetTermVectorsReader() TermVectorsReader
	GetPostingsReader() FieldsProducer
	GetFieldInfos() FieldInfos
	MaxDoc() int
	GetDocValuesReader() DocValuesProducer
	GetNormsReader() NormsProducer
	GetPointsReader() PointsReader
}

type CodecReader interface {
	LeafReader

	// GetFieldsReader
	// Expert: retrieve thread-private StoredFieldsReader
	// lucene.internal
	GetFieldsReader() StoredFieldsReader

	// GetTermVectorsReader
	// Expert: retrieve thread-private TermVectorsReader
	// lucene.internal
	GetTermVectorsReader() TermVectorsReader

	// GetNormsReader
	// Expert: retrieve underlying NormsProducer
	// lucene.internal
	GetNormsReader() NormsProducer

	// GetDocValuesReader
	// Expert: retrieve underlying DocValuesProducer
	// lucene.internal
	GetDocValuesReader() DocValuesProducer

	// GetPostingsReader
	// Expert: retrieve underlying FieldsProducer
	// lucene.internal
	GetPostingsReader() FieldsProducer

	// GetPointsReader
	// Expert: retrieve underlying PointsReader
	// lucene.internal
	GetPointsReader() PointsReader
}
