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

	// GetTokenStreamFromReader 传入io.Reader，生成新的tokenStream对象
	GetTokenStreamFromReader(fieldName string, reader io.Reader) (TokenStream, error)

	// GetTokenStreamFromText 传入string类型，内部缓存buffer复用
	GetTokenStreamFromText(fieldName string, text string) (TokenStream, error)

	// GetPositionIncrementGap Invoked before indexing a IndexableField instance if terms have already been
	// added to that field. This allows custom analyzers to place an automatic position increment gap between
	// IndexbleField instances using the same field name. The default value position increment gap is 0.
	// With a 0 position increment gap and the typical default token position increment of 1, all terms in a field,
	// including across IndexableField instances, are in successive positions, allowing exact PhraseQuery matches,
	// for instance, across IndexableField instance boundaries.
	//
	// fieldName: IndexableField name being indexed.
	// position: increment gap, added to the next token emitted from tokenStream(String, Reader). This value must be >= 0.
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

type ComponentsBuilder interface {
	CreateComponents(fieldName string) *TokenStreamComponents
}

type DefAnalyzer struct {
	builder       ComponentsBuilder
	reuseStrategy ReuseStrategy
	version       *util.Version
	storedValue   any
}

func NewAnalyzer(builder ComponentsBuilder) *DefAnalyzer {
	return &DefAnalyzer{
		builder:       builder,
		reuseStrategy: &GlobalReuseStrategy{},
		version:       util.VersionLast,
	}
}

func (r *DefAnalyzer) Close() error {
	return nil
}

func (r *DefAnalyzer) GetTokenStreamFromText(fieldName string, text string) (TokenStream, error) {
	components := r.reuseStrategy.GetReusableComponents(r, fieldName)

	if components == nil {
		components = r.builder.CreateComponents(fieldName)
		r.reuseStrategy.SetReusableComponents(r, fieldName, components)
	}

	if components.reusableBuffer == nil {
		components.reusableBuffer = new(bytes.Buffer)
	}
	components.reusableBuffer.Reset()
	components.reusableBuffer.WriteString(text)

	strReader := components.reusableBuffer

	components.setReader(strReader)
	return components.GetTokenStream(), nil
}

func (r *DefAnalyzer) initReader(fieldName string, reader io.Reader) io.Reader {
	return reader
}

func (r *DefAnalyzer) GetTokenStreamFromReader(fieldName string, reader io.Reader) (TokenStream, error) {
	components := r.reuseStrategy.GetReusableComponents(r, fieldName)
	if components == nil {
		components = r.builder.CreateComponents(fieldName)
		r.reuseStrategy.SetReusableComponents(r, fieldName, components)
	}
	components.setReader(reader)
	return components.GetTokenStream(), nil
}

func (r *DefAnalyzer) GetPositionIncrementGap(fieldName string) int {
	return 0
}

func (r *DefAnalyzer) GetOffsetGap(fieldName string) int {
	return 1
}

func (r *DefAnalyzer) GetReuseStrategy() ReuseStrategy {
	return r.reuseStrategy
}

func (r *DefAnalyzer) SetVersion(v *util.Version) {
	r.version = v
}

func (r *DefAnalyzer) GetVersion() *util.Version {
	return r.version
}

type ReuseStrategy interface {
	// GetReusableComponents Gets the reusable TokenStreamComponents for the field with the given name.
	// Returns: Reusable TokenStreamComponents for the field, or null if there was no previous components
	// for the field
	// analyzer: 	Analyzer from which to get the reused components.
	//				Use getStoredValue(Analyzer) and setStoredValue(Analyzer, Object) to access the data on the Analyzer.
	// fieldName: Name of the field whose reusable TokenStreamComponents are to be retrieved
	GetReusableComponents(analyzer Analyzer, fieldName string) *TokenStreamComponents

	SetReusableComponents(analyzer Analyzer, fieldName string, components *TokenStreamComponents)
}

type GlobalReuseStrategy struct {
}

func (g *GlobalReuseStrategy) GetReusableComponents(analyzer Analyzer, fieldName string) *TokenStreamComponents {
	switch analyzer.(type) {
	case *DefAnalyzer:
		if components, ok := analyzer.(*DefAnalyzer).storedValue.(*TokenStreamComponents); ok {
			return components
		}
	}
	return nil
}

func (g *GlobalReuseStrategy) SetReusableComponents(analyzer Analyzer, fieldName string, components *TokenStreamComponents) {
	switch analyzer.(type) {
	case *DefAnalyzer:
		analyzer.(*DefAnalyzer).storedValue = components
	}
}

type TokenStreamComponents struct {
	source         func(reader io.Reader)
	sink           TokenStream
	reusableBuffer *bytes.Buffer
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
