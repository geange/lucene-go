package standard

import (
	"github.com/geange/lucene-go/core/analysis"
	"io"
)

var _ analysis.Analyzer = &Analyzer{}

type Analyzer struct {
	*analysis.BaseAnalyzer

	stopWord       *analysis.BaseStopWordAnalyzer
	maxTokenLength int
}

func NewAnalyzer(set *analysis.CharArraySet) *Analyzer {
	analyzer := &Analyzer{
		stopWord:       analysis.NewStopWordAnalyzer(set),
		maxTokenLength: 255,
	}
	analyzer.BaseAnalyzer = analysis.NewBaseAnalyzer(analyzer)
	return analyzer
}

// SetMaxTokenLength Set the max allowed token length. Tokens larger than this will be chopped up at this
// token length and emitted as multiple tokens. If you need to skip such large tokens, you could increase
// this max length, and then use LengthFilter to remove long tokens. The default is DEFAULT_MAX_TOKEN_LENGTH.
func (r *Analyzer) SetMaxTokenLength(length int) {
	r.maxTokenLength = length
}

// GetMaxTokenLength Returns the current maximum token length
// See Also: SetMaxTokenLength
func (r *Analyzer) GetMaxTokenLength() int {
	return r.maxTokenLength
}

func (r *Analyzer) CreateComponents(_ string) *analysis.TokenStreamComponents {
	src := NewTokenizer()
	src.setMaxTokenLength(r.maxTokenLength)
	tok1 := analysis.NewLowerCaseFilter(src)
	tok2 := analysis.NewStopFilter(tok1, r.stopWord.GetStopWordSet())
	return analysis.NewTokenStreamComponents(func(reader io.Reader) {
		src.setMaxTokenLength(r.maxTokenLength)
		_ = src.SetReader(reader)
	}, tok2)
}
