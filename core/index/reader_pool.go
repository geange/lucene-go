package index

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
	"go.uber.org/atomic"
)

// ReaderPool
// Holds shared SegmentReader instances. IndexWriter uses SegmentReaders for 1) applying
// deletes/DV updates, 2) doing merges, 3) handing out a real-time reader. This pool reuses instances
// of the SegmentReaders in all these places if it is in "near real-time mode" (getReader() has been
// called on this instance).
type ReaderPool struct {
	readerMap               map[*SegmentCommitInfo]*ReadersAndUpdates // Map<SegmentCommitInfo,ReadersAndUpdates>
	directory               store.Directory
	originalDirectory       store.Directory
	fieldNumbers            *FieldNumbers
	completedDelGenSupplier func() int64
	segmentInfos            *SegmentInfos
	softDeletesField        string
	// This is a "write once" variable (like the organic dye
	// on a DVD-R that may or may not be heated by a laser and
	// then cooled to permanently record the event): it's
	// false, by default until {@link #enableReaderPooling()}
	// is called for the first time,
	// at which point it's switched to true and never changes
	// back to false.  Once this is true, we hold open and
	// reuse SegmentReader instances internally for applying
	// deletes, doing merges, and reopening near real-time
	// readers.
	// in practice this should be called once the readers are likely
	// to be needed and reused ie if IndexWriter#getReader is called.
	poolReaders bool
	closed      *atomic.Bool
}

func NewReaderPool(directory, originalDirectory store.Directory, segmentInfos *SegmentInfos,
	fieldNumbers *FieldNumbers, completedDelGenSupplier func() int64,
	softDeletesField string, reader *StandardDirectoryReader) (*ReaderPool, error) {

	pool := &ReaderPool{
		readerMap:               map[*SegmentCommitInfo]*ReadersAndUpdates{},
		directory:               directory,
		originalDirectory:       originalDirectory,
		fieldNumbers:            fieldNumbers,
		completedDelGenSupplier: completedDelGenSupplier,
		segmentInfos:            segmentInfos,
		softDeletesField:        softDeletesField,
		poolReaders:             false,
		closed:                  atomic.NewBool(false),
	}
	if reader != nil {
		// Pre-enroll all segment readers into the reader pool; this is necessary so
		// any in-memory NRT live docs are correctly carried over, and so NRT readers
		// pulled from this IW share the same segment reader:
		leaves, err := reader.Leaves()
		if err != nil {
			return nil, err
		}
		if segmentInfos.Size() == len(leaves) {
			return nil, errors.New("leaves size not fit")
		}

		for i := 0; i < len(leaves); i++ {
			leaf := leaves[i]
			segReader := leaf.LeafReader().(*SegmentReader)
			newReader, err := NewSegmentReaderV1(segmentInfos.Info(i), segReader, segReader.GetLiveDocs(),
				segReader.GetHardLiveDocs(), segReader.NumDocs(), true)
			if err != nil {
				return nil, err
			}

			deletes, err := pool.newPendingDeletes(newReader, newReader.GetOriginalSegmentInfo())
			if err != nil {
				return nil, err
			}

			updates, err := NewReadersAndUpdatesV1(segmentInfos.getIndexCreatedVersionMajor(), newReader, deletes)
			if err != nil {
				return nil, err
			}
			pool.readerMap[newReader.GetOriginalSegmentInfo()] = updates
		}
	}
	return pool, nil
}

func (p *ReaderPool) anyDocValuesChanges() bool {
	//for (ReadersAndUpdates rld : readerMap.values()) {
	//	// NOTE: we don't check for pending deletes because deletes carry over in RAM to NRT readers
	//	if (rld.getNumDVUpdates() != 0) {
	//		return true;
	//	}
	//}
	// TODO: fix it
	return false
}

func (p *ReaderPool) newPendingDeletes(reader *SegmentReader, info *SegmentCommitInfo) (PendingDeletes, error) {
	if p.softDeletesField == "" {
		return NewPendingDeletes(reader, info)
	}
	return NewPendingSoftDeletes(p.softDeletesField, reader, info)
}

// Enables reader pooling for this pool. This should be called once the readers in this pool are shared
// with an outside resource like an NRT reader. Once reader pooling is enabled a ReadersAndUpdates will
// be kept around in the reader pool on calling release(ReadersAndUpdates, boolean) until the segment
// get dropped via calls to drop(SegmentCommitInfo) or dropAll() or close(). Reader pooling is disabled
// upon construction but can't be disabled again once it's enabled.
func (p *ReaderPool) enableReaderPooling() {
	p.poolReaders = true
}

// Get Obtain a ReadersAndLiveDocs instance from the readerPool. If create is true,
// you must later call release(ReadersAndUpdates, boolean).
func (p *ReaderPool) Get(info *SegmentCommitInfo, create bool) *ReadersAndUpdates {
	panic("")
}
