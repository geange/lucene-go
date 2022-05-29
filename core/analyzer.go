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

	AnalyzerExt
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

type AnalyzerExt interface {
	// GetPositionIncrementGap Invoked before indexing a IndexableField instance if terms have already been
	// added to that field. This allows custom analyzers to place an automatic position increment gap between
	// IndexbleField instances using the same field name. The default value position increment gap is 0.
	// With a 0 position increment gap and the typical default token position increment of 1, all terms in a field,
	// including across IndexableField instances, are in successive positions, allowing exact PhraseQuery matches,
	// for instance, across IndexableField instance boundaries.
	//
	// Params: fieldName – IndexableField name being indexed.
	// Returns: position increment gap, added to the next token emitted from tokenStream(String, Reader).
	//			This value must be >= 0.
	GetPositionIncrementGap(fieldName string) int

	// GetOffsetGap Just like getPositionIncrementGap, except for Token offsets instead. By default this returns 1.
	// This method is only called if the field produced at least one token for indexing.
	// Params: fieldName – the field just indexed
	// Returns: offset gap, added to the next token emitted from tokenStream(String, Reader). This value must be >= 0.
	GetOffsetGap(fieldName string) int

	// GetReuseStrategy Returns the used Analyzer.ReuseStrategy.
	GetReuseStrategy() ReuseStrategy

	// SetVersion Set the version of Lucene this analyzer should mimic the behavior for analysis.
	SetVersion(v *Version)

	// GetVersion Return the version of Lucene this analyzer will mimic the behavior of for analysis.
	GetVersion() *Version
}

type ReuseStrategy interface {
}
