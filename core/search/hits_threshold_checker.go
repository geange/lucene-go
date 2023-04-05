package search

import (
	"fmt"
	"go.uber.org/atomic"
)

// HitsThresholdChecker
// Used for defining custom algorithms to allow searches to early terminate
type HitsThresholdChecker interface {
	IncrementHitCount()
	ScoreMode() *ScoreMode
	GetHitsThreshold() int
	IsThresholdReached() bool
}

func Create(totalHitsThreshold int) (HitsThresholdChecker, error) {
	return NewLocalHitsThresholdChecker(totalHitsThreshold)
}

// Returns a threshold checker that is based on a shared counter
func createShared(totalHitsThreshold int) (HitsThresholdChecker, error) {
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
		globalHitCount:     atomic.NewInt64(0),
	}, nil
}

func (g *GlobalHitsThresholdChecker) IncrementHitCount() {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalHitsThresholdChecker) ScoreMode() *ScoreMode {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalHitsThresholdChecker) GetHitsThreshold() int {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalHitsThresholdChecker) IsThresholdReached() bool {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (l *LocalHitsThresholdChecker) ScoreMode() *ScoreMode {
	//TODO implement me
	panic("implement me")
}

func (l *LocalHitsThresholdChecker) GetHitsThreshold() int {
	//TODO implement me
	panic("implement me")
}

func (l *LocalHitsThresholdChecker) IsThresholdReached() bool {
	//TODO implement me
	panic("implement me")
}
