package index

import "sync"

// BufferedUpdatesStream
// Tracks the stream of FrozenBufferedUpdates. When DocumentsWriterPerThread flushes,
// its buffered deletes and updates are appended to this stream and immediately resolved
// (to actual docIDs, per segment) using the indexing thread that triggered the flush for concurrency.
// When a merge kicks off, we sync to ensure all resolving packets complete. We also apply to all
// segments when NRT reader is pulled, commit/close is called, or when too many deletes or updates
// are buffered and must be flushed (by RAM usage or by count). Each packet is assigned a generation,
// and each flushed or merged segment is also assigned a generation, so we can track which BufferedDeletes
// packets to apply to any given segment.
type BufferedUpdatesStream struct {
	nextGen          int64
	finishedSegments *FinishedSegments
}

func NewBufferedUpdatesStream() *BufferedUpdatesStream {
	return &BufferedUpdatesStream{
		finishedSegments: NewFinishedSegments(),
	}
}

func (b *BufferedUpdatesStream) GetCompletedDelGen() int64 {
	return b.finishedSegments.GetCompletedDelGen()
}

// FinishedSegments Tracks the contiguous range of packets that have finished resolving.
// We need this because the packets are concurrently resolved, and we can only write to
// disk the contiguous completed packets.
type FinishedSegments struct {
	sync.RWMutex

	// Largest del gen, inclusive, for which all prior packets have finished applying.
	completedDelGen int64

	// This lets us track the "holes" in the current frontier of applying del gens;
	// once the holes are filled in we can advance completedDelGen.
	finishedDelGens map[int64]struct{}
}

func NewFinishedSegments() *FinishedSegments {
	return &FinishedSegments{
		finishedDelGens: map[int64]struct{}{},
	}
}

func (f *FinishedSegments) Clear() {
	f.Lock()
	defer f.Unlock()

	f.completedDelGen = 0
	f.finishedDelGens = map[int64]struct{}{}
}

func (f *FinishedSegments) GetCompletedDelGen() int64 {
	f.RLock()
	defer f.RUnlock()

	return f.completedDelGen
}
