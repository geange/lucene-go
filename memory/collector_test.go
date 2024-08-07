package memory

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"math/rand"
	"testing"

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

var _ index.Scorable = &mockScorable{}

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

func (m *mockScorable) GetChildren() ([]index.ChildScorable, error) {
	//TODO implement me
	panic("implement me")
}
