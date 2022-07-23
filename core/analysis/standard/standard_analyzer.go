package standard

import (
	"github.com/geange/lucene-go/core/analysis"
	"io"
)

type StandardAnalyzer struct {
	*analysis.AnalyzerImp
	*analysis.StopWordAnalyzerBaseImp

	maxTokenLength int
}

func NewAnalyzer(set *analysis.CharArraySet) *StandardAnalyzer {
	analyzer := &StandardAnalyzer{
		StopWordAnalyzerBaseImp: analysis.NewStopWordAnalyzerBaseImp(set),
		maxTokenLength:          255,
	}
	analyzer.AnalyzerImp = analysis.NewAnalyzerImp(analyzer)
	return analyzer
}

// SetMaxTokenLength Set the max allowed token length. Tokens larger than this will be chopped up at this
// token length and emitted as multiple tokens. If you need to skip such large tokens, you could increase
// this max length, and then use LengthFilter to remove long tokens. The default is DEFAULT_MAX_TOKEN_LENGTH.
func (r *StandardAnalyzer) SetMaxTokenLength(length int) {
	r.maxTokenLength = length
}

// GetMaxTokenLength Returns the current maximum token length
// See Also: SetMaxTokenLength
func (r *StandardAnalyzer) GetMaxTokenLength() int {
	return r.maxTokenLength
}

func (r *StandardAnalyzer) CreateComponents(_ string) *analysis.TokenStreamComponents {
	src := NewTokenizer()
	src.setMaxTokenLength(r.maxTokenLength)
	tok1 := analysis.NewLowerCaseFilter(src)
	tok2 := analysis.NewStopFilter(tok1, r.GetStopWordSet())
	return analysis.NewTokenStreamComponents(func(reader io.Reader) {
		src.setMaxTokenLength(r.maxTokenLength)
		src.SetReader(reader)
	}, tok2)
}
