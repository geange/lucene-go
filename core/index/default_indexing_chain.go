package index

import (
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
)

var _ DocConsumer = &DefaultIndexingChain{}

type DefaultIndexingChain struct {
	fieldInfos           *FieldInfosBuilder
	termsHash            *TermsHash
	docValuesBytePool    *util.ByteBlockPool
	storedFieldsConsumer *StoredFieldsConsumer
	termVectorsWriter    *TermVectorsConsumer

	// NOTE: I tried using Hash Map<String,PerField>
	// but it was ~2% slower on Wiki and Geonames with Java
	// 1.7.0_25:
	fieldHash []PerField
	hashMask  int

	totalFieldCount int
	nextFieldGen    int64

	fields []PerField

	infoStream io.Writer

	indexWriterConfig *LiveIndexWriterConfig

	indexCreatedVersionMajor int

	hasHitAbortingException bool
}

func (d *DefaultIndexingChain) ProcessDocument(docId int, document []types.IndexableField) error {
	//TODO implement me
	panic("implement me")
}

func (d *DefaultIndexingChain) Flush(state *SegmentWriteState) (DocMap, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DefaultIndexingChain) Abort() error {
	//TODO implement me
	panic("implement me")
}

func (d *DefaultIndexingChain) GetHasDocValues(field string) DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

type PerField struct {
	indexCreatedVersionMajor int
	fieldInfo                *types.FieldInfo
	similarity               Similarity
	invertState              *FieldInvertState
	termsHashPerField        *TermsHashPerField
	docValuesWriter          DocValuesWriter
	pointValuesWriter        *PointValuesWriter

	// We use this to know when a PerField is seen for the first time in the current document.
	fieldGen int64

	// Used by the hash table
	next *PerField

	norms NormValuesWriter

	tokenStream analysis.TokenStream

	infoStream io.Writer

	analyzer analysis.Analyzer
}
