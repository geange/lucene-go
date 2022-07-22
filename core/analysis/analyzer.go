package analysis

import (
	"bytes"
	"github.com/geange/lucene-go/core/util"
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
	SetVersion(v *util.Version)

	// GetVersion Return the version of Lucene this analyzer will mimic the behavior of for analysis.
	GetVersion() *util.Version
}

type AnalyzerDefault struct {
	reuseStrategy ReuseStrategy
	version       *util.Version
}

func NewAnalyzerDefault() *AnalyzerDefault {
	return &AnalyzerDefault{
		version: util.VersionLast,
	}
}

func (*AnalyzerDefault) GetPositionIncrementGap(fieldName string) int {
	return 0
}

func (*AnalyzerDefault) GetOffsetGap(fieldName string) int {
	return 1
}

func (r *AnalyzerDefault) SetVersion(v *util.Version) {
	r.version = v
}

func (r *AnalyzerDefault) GetVersion() *util.Version {
	return r.version
}

type AnalyzerPLG interface {
	CreateComponents(fieldName string) *TokenStreamComponents
}

type AnalyzerExtra interface {
	CreateComponents(fieldName string) *TokenStreamComponents
}

type AnalyzerImp struct {
	AnalyzerExtra

	reuseStrategy ReuseStrategy
	version       *util.Version

	storedValue interface{}
}

func NewAnalyzerImp(plg AnalyzerExtra) *AnalyzerImp {
	return &AnalyzerImp{
		AnalyzerExtra: plg,
		reuseStrategy: &GlobalReuseStrategy{},
		version:       util.VersionLast,
	}
}

func (r *AnalyzerImp) Close() error {
	return nil
}

func (r *AnalyzerImp) TokenStreamByString(fieldName string, text string) (TokenStream, error) {
	components := r.reuseStrategy.GetReusableComponents(r, fieldName)

	if components == nil {
		components = r.CreateComponents(fieldName)
		r.reuseStrategy.SetReusableComponents(r, fieldName, components)
	}

	if components.reusableStringReader == nil {
		components.reusableStringReader = bytes.NewBufferString(text)
	}

	strReader := components.reusableStringReader

	components.setReader(strReader)
	return components.GetTokenStream(), nil
}

func (r *AnalyzerImp) initReader(fieldName string, reader io.Reader) io.Reader {
	return reader
}

func (r *AnalyzerImp) TokenStreamByReader(fieldName string, reader io.Reader) (TokenStream, error) {
	components := r.reuseStrategy.GetReusableComponents(r, fieldName)
	if components == nil {
		components = r.CreateComponents(fieldName)
		r.reuseStrategy.SetReusableComponents(r, fieldName, components)
	}
	components.setReader(reader)
	return components.GetTokenStream(), nil
}

func (r *AnalyzerImp) GetPositionIncrementGap(fieldName string) int {
	return 0
}

func (r *AnalyzerImp) GetOffsetGap(fieldName string) int {
	return 1
}

func (r *AnalyzerImp) GetReuseStrategy() ReuseStrategy {
	return r.reuseStrategy
}

func (r *AnalyzerImp) SetVersion(v *util.Version) {
	r.version = v
}

func (r *AnalyzerImp) GetVersion() *util.Version {
	return r.version
}

type ReuseStrategy interface {
	// GetReusableComponents Gets the reusable TokenStreamComponents for the field with the given name.
	// Params: 	analyzer – Analyzer from which to get the reused components. Use getStoredValue(Analyzer)
	//			and setStoredValue(Analyzer, Object) to access the data on the Analyzer.
	//			fieldName – Name of the field whose reusable TokenStreamComponents are to be retrieved
	// Returns: Reusable TokenStreamComponents for the field, or null if there was no previous components
	//			for the field
	GetReusableComponents(analyzer Analyzer, fieldName string) *TokenStreamComponents

	SetReusableComponents(analyzer Analyzer, fieldName string, components *TokenStreamComponents)
}

type GlobalReuseStrategy struct {
}

func (g *GlobalReuseStrategy) GetReusableComponents(analyzer Analyzer, fieldName string) *TokenStreamComponents {
	switch analyzer.(type) {
	case *AnalyzerImp:
		if components, ok := analyzer.(*AnalyzerImp).storedValue.(*TokenStreamComponents); ok {
			return components
		}
	}
	return nil
}

func (g *GlobalReuseStrategy) SetReusableComponents(analyzer Analyzer, fieldName string, components *TokenStreamComponents) {
	switch analyzer.(type) {
	case *AnalyzerImp:
		analyzer.(*AnalyzerImp).storedValue = components
	}
}

type TokenStreamComponents struct {
	source               func(reader io.Reader)
	sink                 TokenStream
	reusableStringReader *bytes.Buffer
}

func NewTokenStreamComponents(source func(reader io.Reader), result TokenStream) *TokenStreamComponents {
	return &TokenStreamComponents{
		source: source,
		sink:   result,
	}
}

func (r *TokenStreamComponents) setReader(reader io.Reader) {
	r.source(reader)
}

func (r *TokenStreamComponents) GetTokenStream() TokenStream {
	return r.sink
}

func (r *TokenStreamComponents) GetSource() func(reader io.Reader) {
	return r.source
}
