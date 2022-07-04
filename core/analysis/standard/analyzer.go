package standard

import (
	"github.com/geange/lucene-go/core/analysis"
	"io"
)

type Analyzer struct {
	*analysis.StopWordAnalyzerBaseIMP

	maxTokenLength int
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
	tok := analysis.NewLowerCaseFilter(src)
	return analysis.NewTokenStreamComponents(func(reader io.Reader) {
		src.setMaxTokenLength(r.maxTokenLength)
		src.SetReader(reader)
	}, tok)
}
