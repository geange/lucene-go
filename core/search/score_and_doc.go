package search

var _ Scorable = &ScoreAndDoc{}

type ScoreAndDoc struct {
	*ScorableDefault

	score float32
	doc   int
}

func NewScoreAndDoc() *ScoreAndDoc {
	return &ScoreAndDoc{ScorableDefault: &ScorableDefault{}}
}

func (s *ScoreAndDoc) Score() (float32, error) {
	return s.score, nil
}

func (s *ScoreAndDoc) DocID() int {
	return s.doc
}
