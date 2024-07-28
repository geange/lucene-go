package index

import (
	"sync/atomic"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/util/hash"
)

// BufferedUpdates
// Holds buffered deletes and updates, by docID, term or query for a single segment.
// This is used to hold buffered pending deletes and updates against the to-be-flushed segment.
// Once the deletes and updates are pushed (on Flush in DocumentsWriter), they are converted to a
// FrozenBufferedUpdates instance and pushed to the BufferedUpdatesStream.
//
// NOTE: instances of this class are accessed either via a private
// instance on DocumentWriterPerThread, or via sync'd code by
// DocumentsWriterDeleteQueue
type BufferedUpdates struct {
	numTermDeletes  *atomic.Int64
	numFieldUpdates *atomic.Int64
	deleteTerms     *treemap.Map[Term, int]
	fieldUpdates    map[string]*FieldUpdatesBuffer
	gen             int64
	segmentName     string
	deleteQueries   *treemap.Map[Query, int]
}

func (b *BufferedUpdates) GetNumFieldUpdates() int64 {
	return b.numFieldUpdates.Load()
}

type bufferedUpdatesOption struct {
	segmentName string
}

func WithSegmentName(segmentName string) BufferedUpdatesOption {
	return func(o *bufferedUpdatesOption) {
		o.segmentName = segmentName
	}
}

type BufferedUpdatesOption func(*bufferedUpdatesOption)

func NewBufferedUpdates(options ...BufferedUpdatesOption) *BufferedUpdates {
	opt := &bufferedUpdatesOption{}
	for _, fn := range options {
		fn(opt)
	}

	return &BufferedUpdates{
		numTermDeletes:  new(atomic.Int64),
		numFieldUpdates: new(atomic.Int64),
		deleteTerms:     treemap.NewWith[Term, int](TermCompare),
		segmentName:     opt.segmentName,
		deleteQueries:   treemap.NewWith[Query, int](hash.Compare[Query]),
	}
}

func (b *BufferedUpdates) AddTerm(term Term, docIDUpto int) {
	current, ok := b.deleteTerms.Get(term)
	if ok {
		if current > docIDUpto {
			// Only record the new number if it's greater than the
			// current one.  This is important because if multiple
			// threads are replacing the same doc at nearly the
			// same time, it's possible that one thread that got a
			// higher docID is scheduled before the other
			// threads.  If we blindly replace than we can
			// incorrectly get both docs indexed.
			return
		}
	}

	b.deleteTerms.Put(term, docIDUpto)
	// note that if current != null then it means there's already a buffered
	// delete on that term, therefore we seem to over-count. this over-counting
	// is done to respect IndexWriterConfig.setMaxBufferedDeleteTerms.
	b.numTermDeletes.Add(1)
}

func (b *BufferedUpdates) AddNumericUpdate(update *NumericDocValuesUpdate, docIDUpto int) error {
	field := update.GetField()
	buffer, ok := b.fieldUpdates[field]
	if !ok {
		newBuffer := NewNumberFieldUpdatesBuffer(update, docIDUpto)
		b.fieldUpdates[field] = newBuffer
		buffer = newBuffer
	}

	if update.HasValue() {
		err := buffer.addUpdateInt(update.term, update.GetValue(), docIDUpto)
		if err != nil {
			return err
		}
	} else {
		err := buffer.addNoValue(update.term, docIDUpto)
		if err != nil {
			return err
		}
	}
	b.numFieldUpdates.Add(1)
	return nil
}

func (b *BufferedUpdates) AddBinaryUpdate(update *BinaryDocValuesUpdate, docIDUpto int) error {
	field := update.GetField()
	buffer, ok := b.fieldUpdates[field]
	if !ok {
		newBuffer := NewBinaryFieldUpdatesBuffer(update, docIDUpto)
		b.fieldUpdates[field] = newBuffer
		buffer = newBuffer
	}

	if update.HasValue() {
		err := buffer.addUpdateBytes(update.term, update.GetValue(), docIDUpto)
		if err != nil {
			return err
		}
	} else {
		err := buffer.addNoValue(update.term, docIDUpto)
		if err != nil {
			return err
		}
	}
	b.numFieldUpdates.Add(1)
	return nil
}

func (b *BufferedUpdates) ClearDeleteTerms() {
	b.numTermDeletes.Store(0)
	//b.termsBytesUsed.addAndGet(-termsBytesUsed.get());
	b.deleteTerms.Clear()
}

func (b *BufferedUpdates) Clear() {
	b.deleteTerms.Clear()
	b.deleteQueries.Clear()
	b.numTermDeletes.Store(0)
	b.numFieldUpdates.Store(0)
	clear(b.fieldUpdates)
}

func (b *BufferedUpdates) Any() bool {
	return b.deleteTerms.Size() > 0 ||
		b.deleteQueries.Size() > 0 ||
		b.numFieldUpdates.Load() > 0
}

func (b *BufferedUpdates) GetDeleteQueries() *treemap.Map[Query, int] {
	return b.deleteQueries
}
