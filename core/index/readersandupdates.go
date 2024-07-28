package index

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"sync/atomic"

	"github.com/geange/lucene-go/core/store"
)

// ReadersAndUpdates
// Used by IndexWriter to hold open SegmentReaders (for searching or merging), plus pending deletes and updates,
// for a given segment
type ReadersAndUpdates struct {
	// Not final because we replace (clone) when we need to
	// change it and it's been shared:
	info *index.SegmentCommitInfo

	// Tracks how many consumers are using this instance:
	refCount *atomic.Int64

	// Set once (null, and then maybe set, and never set again):
	reader *SegmentReader

	// How many further deletions we've done against
	// liveDocs vs when we loaded it or last wrote it:
	pendingDeletes PendingDeletes

	// the major version this index was created with
	indexCreatedVersionMajor int

	// Indicates whether this segment is currently being merged. While a segment
	// is merging, all field updates are also registered in the
	// mergingNumericUpdates map. Also, calls to writeFieldUpdates merge the
	// updates with mergingNumericUpdates.
	// That way, when the segment is done merging, IndexWriter can apply the
	// updates on the merged segment too.
	isMerging bool

	// Holds resolved (to docIDs) doc values updates that have not yet been
	// written to the index
	pendingDVUpdates map[string][]DocValuesFieldUpdates

	// Holds resolved (to docIDs) doc values updates that were resolved while
	// this segment was being merged; at the end of the merge we carry over
	// these updates (remapping their docIDs) to the newly merged segment
	mergingDVUpdates map[string][]DocValuesFieldUpdates

	// Only set if there are doc values updates against this segment, and the index is sorted:
	sortMap DocMap
}

func NewReadersAndUpdates(indexCreatedVersionMajor int,
	info *index.SegmentCommitInfo, pendingDeletes PendingDeletes) *ReadersAndUpdates {

	return &ReadersAndUpdates{
		info:                     info,
		refCount:                 new(atomic.Int64),
		reader:                   nil,
		pendingDeletes:           pendingDeletes,
		indexCreatedVersionMajor: indexCreatedVersionMajor,
		isMerging:                false,
		pendingDVUpdates:         map[string][]DocValuesFieldUpdates{},
		mergingDVUpdates:         map[string][]DocValuesFieldUpdates{},
	}
}

// NewReadersAndUpdates
// Init from a previously opened SegmentReader.
// NOTE: steals incoming ref from reader.
func (s *SegmentReader) NewReadersAndUpdates(indexCreatedVersionMajor int, pendingDeletes PendingDeletes) (*ReadersAndUpdates, error) {
	updates := NewReadersAndUpdates(indexCreatedVersionMajor, s.GetOriginalSegmentInfo(), pendingDeletes)
	updates.reader = s
	if err := pendingDeletes.OnNewReader(s, updates.info); err != nil {
		return nil, err
	}
	return updates, nil
}

func (r *ReadersAndUpdates) IncRef() {
	r.refCount.Add(1)
}

func (r *ReadersAndUpdates) DecRef() {
	r.refCount.Add(-1)
}

func (r *ReadersAndUpdates) RefCount() int64 {
	return r.refCount.Load()
}

func (r *ReadersAndUpdates) GetDelCount() int {
	return r.pendingDeletes.GetDelCount()
}

// AddDVUpdate
// Adds a new resolved (meaning it maps docIDs to new values) doc values packet.
// We buffer these in RAM and write to disk when too much RAM is used or when a merge needs
// to kick off, or a commit/refresh.
func (r *ReadersAndUpdates) AddDVUpdate(update DocValuesFieldUpdates) error {
	if update.GetFinished() == false {
		return errors.New("call finish first")
	}

	field := update.Field()

	if _, ok := r.pendingDVUpdates[field]; !ok {
		r.pendingDVUpdates[field] = []DocValuesFieldUpdates{}
	}

	r.pendingDVUpdates[field] = append(r.pendingDVUpdates[field], update)

	if r.isMerging {
		_, ok := r.mergingDVUpdates[field]
		if !ok {
			r.mergingDVUpdates[field] = []DocValuesFieldUpdates{}
		}
		r.mergingDVUpdates[field] = append(r.mergingDVUpdates[field], update)
	}
	return nil
}

func (r *ReadersAndUpdates) GetNumDVUpdates() int {
	count := 0
	for _, updates := range r.pendingDVUpdates {
		count += len(updates)
	}
	return count
}

func (r *ReadersAndUpdates) GetReader(ctx context.Context, ioContext *store.IOContext) (*SegmentReader, error) {
	if r.reader == nil {
		// We steal returned ref:
		reader, err := NewSegmentReader(ctx, r.info, r.indexCreatedVersionMajor, ioContext)
		if err != nil {
			return nil, err
		}
		r.reader = reader
		err = r.pendingDeletes.OnNewReader(r.reader, r.info)
		if err != nil {
			return nil, err
		}
	}

	// Ref for caller
	if err := r.reader.IncRef(); err != nil {
		return nil, err
	}
	return r.reader, nil
}

func (r *ReadersAndUpdates) Release(sr *SegmentReader) error {
	return sr.DecRef()
}

func (r *ReadersAndUpdates) writeLiveDocs(directory store.Directory) (bool, error) {
	panic("")
}

func (r *ReadersAndUpdates) writeFieldUpdates(directory store.Directory, numbers *FieldNumbers, supplier int64) (bool, error) {
	panic("")
}

func (r *ReadersAndUpdates) IsFullyDeleted() (bool, error) {
	return r.pendingDeletes.IsFullyDeleted(nil, r.getLatestReader)
}

func (r *ReadersAndUpdates) getLatestReader() CodecReader {
	panic("")
}
