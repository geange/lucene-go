package search

import (
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"
	"math"
	"sync/atomic"
)

// HitsThresholdChecker
// Used for defining custom algorithms to allow searches to early terminate
type HitsThresholdChecker interface {
	IncrementHitCount()
	ScoreMode() index.ScoreMode
	GetHitsThreshold() int
	IsThresholdReached() bool
}

func HitsThresholdCheckerCreate(totalHitsThreshold int) (HitsThresholdChecker, error) {
	return NewLocalHitsThresholdChecker(totalHitsThreshold)
}

// HitsThresholdCheckerCreateShared
// Returns a threshold checker that is based on a shared counter
func HitsThresholdCheckerCreateShared(totalHitsThreshold int) (HitsThresholdChecker, error) {
	return NewGlobalHitsThresholdChecker(totalHitsThreshold)
}

var _ HitsThresholdChecker = &GlobalHitsThresholdChecker{}

// GlobalHitsThresholdChecker
// Implementation of HitsThresholdChecker which allows global hit counting
type GlobalHitsThresholdChecker struct {
	totalHitsThreshold int
	globalHitCount     *atomic.Int64
}

func NewGlobalHitsThresholdChecker(totalHitsThreshold int) (*GlobalHitsThresholdChecker, error) {
	return &GlobalHitsThresholdChecker{
		totalHitsThreshold: totalHitsThreshold,
		globalHitCount:     new(atomic.Int64),
	}, nil
}

func (g *GlobalHitsThresholdChecker) IncrementHitCount() {
	g.globalHitCount.Add(1)
}

func (g *GlobalHitsThresholdChecker) ScoreMode() index.ScoreMode {
	if g.totalHitsThreshold == math.MaxInt32 {
		return COMPLETE
	}
	return TOP_SCORES
}

func (g *GlobalHitsThresholdChecker) GetHitsThreshold() int {
	return g.totalHitsThreshold
}

func (g *GlobalHitsThresholdChecker) IsThresholdReached() bool {
	return g.globalHitCount.Load() > int64(g.totalHitsThreshold)
}

var _ HitsThresholdChecker = &LocalHitsThresholdChecker{}

// LocalHitsThresholdChecker
// Default implementation of HitsThresholdChecker to be used for single threaded execution
type LocalHitsThresholdChecker struct {
	totalHitsThreshold int
	hitCount           int
}

func NewLocalHitsThresholdChecker(totalHitsThreshold int) (*LocalHitsThresholdChecker, error) {
	if totalHitsThreshold < 0 {
		return nil, fmt.Errorf("totalHitsThreshold must be >= 0, got %d", totalHitsThreshold)
	}
	return &LocalHitsThresholdChecker{
		totalHitsThreshold: totalHitsThreshold,
	}, nil
}

func (l *LocalHitsThresholdChecker) IncrementHitCount() {
	l.hitCount++
}

func (l *LocalHitsThresholdChecker) ScoreMode() index.ScoreMode {
	if l.totalHitsThreshold == math.MaxInt32 {
		return COMPLETE
	}
	return TOP_SCORES
}

func (l *LocalHitsThresholdChecker) GetHitsThreshold() int {
	return l.totalHitsThreshold
}

func (l *LocalHitsThresholdChecker) IsThresholdReached() bool {
	return l.hitCount > l.totalHitsThreshold
}
