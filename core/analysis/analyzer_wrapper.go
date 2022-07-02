package analysis

// AnalyzerWrapper Extension to Analyzer suitable for Analyzers which wrap other Analyzers.
// getWrappedAnalyzer(String) allows the Analyzer to wrap multiple Analyzers which are selected on a per field basis.
// wrapComponents(String, Analyzer.TokenStreamComponents) allows the TokenStreamComponents of the wrapped Analyzer
// to then be wrapped (such as adding a new TokenFilter to form new TokenStreamComponents.
// wrapReader(String, Reader) allows the Reader of the wrapped Analyzer to then be wrapped (such as adding a
// new CharFilter.
// Important: If you do not want to wrap the TokenStream using wrapComponents(String, Analyzer.TokenStreamComponents)
// or the Reader using wrapReader(String, Reader) and just delegate to other analyzers (like by field name),
// use DelegatingAnalyzerWrapper as superclass!
// Since: 4.0.0
// See Also: DelegatingAnalyzerWrapper
type AnalyzerWrapper interface {
	Analyzer
}
