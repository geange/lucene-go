package memory

import (
	"github.com/geange/lucene-go/codecs"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ codecs.FieldsConsumer = &FSTTermsWriter{}

const (
	TERMS_EXTENSION       = "tfp"
	TERMS_CODEC_NAME      = "FSTTerms"
	TERMS_VERSION_START   = 2
	TERMS_VERSION_CURRENT = TERMS_VERSION_START
)

type FSTTermsWriter struct {
	postingsWriter codecs.PostingsWriterBase
	fieldInfos     *index.FieldInfos
	out            store.IndexOutput
	maxDoc         int
	fields         []FieldMetaData
}

type FieldMetaData struct {
	fieldInfo        *index.FieldInfo
	numTerms         int64
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	dict             FST
}
