package core

import (
	"io"
)

// An Analyzer builds TokenStreams, which analyze text. It thus represents a policy for
// extracting index terms from text.
//
// For some concrete implementations bundled with Lucene, look in the analysis modules:
// * Common: Analyzers for indexing content in different languages and domains.
// * ICU: Exposes functionality from ICU to Apache Lucene.
// * Kuromoji: Morphological analyzer for Japanese text.
// * Morfologik: Dictionary-driven lemmatization for the Polish language.
// * Phonetic: Analysis for indexing phonetic signatures (for sounds-alike search).
// * Smart Chinese: Analyzer for Simplified Chinese, which indexes words.
// * Stempel: Algorithmic Stemmer for the Polish Language.
//
// Analyzer 用来构建 TokenStream 用于分析文本。Analyzer 代表了从文本中提取索引项(index term)的策略
type Analyzer interface {
	io.Closer

	TokenStreamByReader(fieldName string, reader io.Reader) (TokenStream, error)
	TokenStreamByString(fieldName string, text string) (TokenStream, error)
}

//func NewTokenStream[T io.Reader | string](fieldName string, data T) (TokenStream, error) {
//	switch data.(type) {
//	case io.Reader:
//		panic("")
//	case string:
//		panic("")
//	default:
//		return nil, errors.New("")
//	}
//}
