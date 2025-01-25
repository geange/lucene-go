package index

import (
	"errors"
	"sync/atomic"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

// ReaderPool
// Holds shared SegmentReader instances. IndexWriter uses SegmentReaders for 1) applying
// deletes/DV updates, 2) doing merges, 3) handing out a real-time reader. This pool reuses instances
// of the SegmentReaders in all these places if it is in "near real-time mode" (getReader() has been
// called on this instance).
type ReaderPool struct {
	readerMap               map[index.SegmentCommitInfo]*ReadersAndUpdates // Map<SegmentCommitInfo,ReadersAndUpdates>
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
		readerMap:               map[index.SegmentCommitInfo]*ReadersAndUpdates{},
		directory:               directory,
		originalDirectory:       originalDirectory,
		fieldNumbers:            fieldNumbers,
		completedDelGenSupplier: completedDelGenSupplier,
		segmentInfos:            segmentInfos,
		softDeletesField:        softDeletesField,
		poolReaders:             false,
		closed:                  new(atomic.Bool),
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
			newReader, err := segReader.New(segmentInfos.Info(i), segReader.GetLiveDocs(),
				segReader.GetHardLiveDocs(), segReader.NumDocs(), true)
			if err != nil {
				return nil, err
			}

			deletes := pool.newPendingDeletesV1(newReader, newReader.GetOriginalSegmentInfo())
			updates, err := newReader.NewReadersAndUpdates(segmentInfos.getIndexCreatedVersionMajor(), deletes)
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

func (p *ReaderPool) newPendingDeletes(info index.SegmentCommitInfo) PendingDeletes {

	if p.softDeletesField == "" {
		return NewPendingDeletesV1(info)
	}
	return NewPendingSoftDeletes(p.softDeletesField, info)
}

func (p *ReaderPool) newPendingDeletesV1(reader *SegmentReader, info index.SegmentCommitInfo) PendingDeletes {
	if p.softDeletesField == "" {
		return NewPendingDeletes(reader, info)
	}
	return NewPendingSoftDeletesV1(p.softDeletesField, reader, info)
}

// Enables reader pooling for this pool. This should be called once the readers in this pool are shared
// with an outside resource like an NRT reader. Once reader pooling is enabled a ReadersAndUpdates will
// be kept around in the reader pool on calling release(ReadersAndUpdates, boolean) until the segment
// get dropped via calls to drop(SegmentCommitInfo) or dropAll() or close(). IndexReader pooling is disabled
// upon construction but can't be disabled again once it's enabled.
func (p *ReaderPool) enableReaderPooling() {
	p.poolReaders = true
}

// Get
// Obtain a ReadersAndLiveDocs instance from the readerPool. If create is true,
// you must later call release(ReadersAndUpdates, boolean).
func (p *ReaderPool) Get(info index.SegmentCommitInfo, create bool) (*ReadersAndUpdates, error) {
	if p.closed.Load() {
		return nil, errors.New("ReaderPool is already closed")
	}

	rld, ok := p.readerMap[info]
	if !ok {
		if !create {
			return nil, nil
		}
		rld = NewReadersAndUpdates(p.segmentInfos.getIndexCreatedVersionMajor(), info, p.newPendingDeletes(info))
		// Steal initial reference:
		p.readerMap[info] = rld
	}

	if create {
		// Return ref to caller:
		rld.IncRef()
	}

	return rld, nil
}

func (p *ReaderPool) commit(infos *SegmentInfos) (bool, error) {
	atLeastOneChange := false
	for _, segment := range infos.segments {
		rld, ok := p.readerMap[segment]
		if !ok {
			continue
		}

		changed, err := rld.writeLiveDocs(p.directory)
		if err != nil {
			return false, err
		}
		updates, err := rld.writeFieldUpdates(p.directory, p.fieldNumbers, p.completedDelGenSupplier())
		if err != nil {
			return false, err
		}
		changed = changed || updates

		if changed {
			// Make sure we only write del docs for a live segment:
			//assert assertInfoIsLive(info);

			// Must checkpoint because we just
			// created new _X_N.del and field updates files;
			// don't call IW.checkpoint because that also
			// increments SIS.version, which we do not want to
			// do here: it was done previously (after we
			// invoked BDS.applyDeletes), whereas here all we
			// did was move the state to disk:
			atLeastOneChange = true
		}
	}
	return atLeastOneChange, nil
}

func (p *ReaderPool) writeAllDocValuesUpdates() (bool, error) {
	panic("")
}
