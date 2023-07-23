package types

import (
	"github.com/geange/lucene-go/core/analysis"
)

// IndexableField Represents a single field for indexing. IndexWriter consumes
// []IndexableField as a document.
// IndexAbleField代表一个可以被索引的field，每一个Document都是由多个IndexAbleField组成
type IndexableField interface {
	// Name 获取Field name
	Name() string

	// FieldType 获取field的属性
	FieldType() IndexableFieldType

	// TokenStream Creates the TokenStream used for indexing this field. If appropriate, implementations should
	// use the given Analyzer to create the TokenStreams.
	// Params: 	analyzer – Analyzer that should be used to create the TokenStreams from
	//			reuse – TokenStream for a previous instance of this field name. This allows custom field types
	//			(like StringField and NumericField) that do not use the analyzer to still have good performance.
	//			Note: the passed-in types may be inappropriate, for example if you mix up different types of Fields
	//			for the same field name. So it's the responsibility of the implementation to check.
	// Returns: TokenStream value for indexing the document. Should always return a non-null value if the field
	//			is to be indexed
	TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (analysis.TokenStream, error)

	// Value 内容信息
	Value() any
}

type IndexableFieldIterator interface {
	Iterator() func() IndexableField
}
