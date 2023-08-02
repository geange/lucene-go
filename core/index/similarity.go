package index

import (
	"fmt"
	"github.com/geange/lucene-go/core/types"
)

type SimScorerSPI interface {
	Score(freq float64, norm int64) float64
}

type BaseSimScorer struct {
	SimScorerSPI
}

func NewBaseSimScorer(simScorerSPI SimScorerSPI) *BaseSimScorer {
	return &BaseSimScorer{SimScorerSPI: simScorerSPI}
}

func (s *BaseSimScorer) Explain(freq *types.Explanation, norm int64) (*types.Explanation, error) {
	return types.ExplanationMatch(
		s.Score(freq.GetValue().(float64), norm),
		fmt.Sprintf(`score(freq="%v"), with freq of:`, freq.GetValue()),
		freq), nil
}
