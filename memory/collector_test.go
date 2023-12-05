package memory

import (
	"context"
	"math/rand"
	"testing"

	"github.com/geange/lucene-go/core/search"
	"github.com/stretchr/testify/assert"
)

func TestNewSimpleCollector(t *testing.T) {
	scores := make([]float64, 1)
	collector := newSimpleCollector(scores)

	score := rand.Float64()
	err := collector.SetScorer(&mockScorable{score: score})
	assert.Nil(t, err)

	err = collector.Collect(context.Background(), 1)
	assert.Nil(t, err)

	assert.InDelta(t, score, scores[0], 0.0000001)
}

var _ search.Scorable = &mockScorable{}

type mockScorable struct {
	score float64
}

func (m *mockScorable) Score() (float64, error) {
	return m.score, nil
}

func (m *mockScorable) SmoothingScore(docId int) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockScorable) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (m *mockScorable) SetMinCompetitiveScore(minScore float64) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockScorable) GetChildren() ([]search.ChildScorable, error) {
	//TODO implement me
	panic("implement me")
}
