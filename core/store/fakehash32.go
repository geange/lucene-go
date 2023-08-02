package store

import "hash"

type fakeHash32 struct {
	hash.Hash32
}

func NewFakeHash32() hash.Hash32 {
	return fakeHash32{}
}

func (fakeHash32) Write(p []byte) (int, error) { return len(p), nil }
func (fakeHash32) Sum32() uint32               { return 0 }
