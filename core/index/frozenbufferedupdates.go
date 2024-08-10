package index

import (
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"sync"
	"sync/atomic"
	"unsafe"
)

// FrozenBufferedUpdates
// Holds buffered deletes and updates by term or query, once pushed. Pushed deletes/updates are write-once,
// so we shift to more memory efficient data structure to hold them. We don't hold docIDs because these are
// applied on flush.
type FrozenBufferedUpdates struct {
	sync.Mutex

	// Terms, in sorted order:
	deleteTerms *PrefixCodedTerms

	// Parallel array of deleted query, and the docIDUpto for each
	deleteQueries     []index.Query
	deleteQueryLimits []int

	// Counts down once all deletes/ updates have been applied
	fieldUpdates map[string]*index.FieldUpdatesBuffer

	// How many total documents were deleted/updated.
	totalDelCount     int
	fieldUpdatesCount int

	numTermDeletes int

	delGen int64 // assigned by BufferedUpdatesStream once pushed

	privateSegment *index.SegmentCommitInfo // non-null iff this frozen packet represents
}

// NewFrozenBufferedUpdates
// TODO: fix it
func NewFrozenBufferedUpdates(updates *index.BufferedUpdates, privateSegment index.SegmentCommitInfo) *FrozenBufferedUpdates {
	return &FrozenBufferedUpdates{
		delGen: -1,
	}
}

type Locker struct {
	sync.Mutex
	isLocked *atomic.Bool
}

type mutex struct {
	state int32
	sema  uint32
}

// Returns true if this buffered updates instance was already applied
func (f *FrozenBufferedUpdates) isApplied() bool {
	mux := (*mutex)(unsafe.Pointer(&(f.Mutex)))
	return mux.state == 0
}

// Apply
// Applies pending delete-by-term, delete-by-query and doc values updates to all segments in the index,
// returning the number of new deleted or updated documents.
func (f *FrozenBufferedUpdates) Apply(segStates []*SegmentState) (int, error) {

	if f.delGen == -1 {
		// we were not yet pushed
		return 0, errors.New("gen is not yet set; call BufferedUpdatesStream.push first")
	}

	termDeletesCount, err := f.applyTermDeletes(segStates)
	if err != nil {
		return 0, err
	}
	f.totalDelCount += termDeletesCount

	queryDeletesCount, err := f.applyQueryDeletes(segStates)
	if err != nil {
		return 0, err
	}
	f.totalDelCount += queryDeletesCount

	updatesCount, err := f.applyDocValuesUpdates(segStates)
	if err != nil {
		return 0, err
	}
	f.totalDelCount += updatesCount

	return f.totalDelCount, nil

}

func (f *FrozenBufferedUpdates) applyTermDeletes(segStates []*SegmentState) (int, error) {
	if f.deleteTerms.Size() == 0 {
		return 0, nil
	}

	//delCount := 0

	for _, segState := range segStates {
		if segState.delGen > f.delGen {
			// our deletes don't apply to this segment
			continue
		}

		if segState.rld.RefCount() == 1 {
			// This means we are the only remaining reference to this segment, meaning
			// it was merged away while we were running, so we can safely skip running
			// because we will run on the newly merged segment next:
			continue
		}

		// iter := f.deleteTerms.Iterator();
		//BytesRef delTerm;
		//TermDocsIterator termDocsIterator = new TermDocsIterator(segState.reader, true);
	}
	panic("")
}

func (f *FrozenBufferedUpdates) applyQueryDeletes(segStates []*SegmentState) (int, error) {
	panic("")
}

func (f *FrozenBufferedUpdates) applyDocValuesUpdates(segStates []*SegmentState) (int, error) {
	panic("")
}

func (f *FrozenBufferedUpdates) Any() bool {
	return f.deleteTerms.Size() > 0 || len(f.deleteQueries) > 0 || f.fieldUpdatesCount > 0
}

type TermDocsIterator struct {
}
