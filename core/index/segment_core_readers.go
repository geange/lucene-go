package index

import (
	"github.com/geange/lucene-go/core/store"
	"go.uber.org/atomic"
)

// SegmentCoreReaders
// Holds core readers that are shared (unchanged) when SegmentReader is cloned or reopened
type SegmentCoreReaders struct {

	// Counts how many other readers share the core objects
	// (freqStream, proxStream, tis, etc.) of this reader;
	// when coreRef drops to 0, these core objects may be
	// closed.  A given instance of SegmentReader may be
	// closed, even though it shares core objects with other
	// SegmentReaders:

	ref                   *atomic.Int64
	fields                FieldsProducer
	normsProducer         NormsProducer
	fieldsReaderOrig      StoredFieldsReader
	termVectorsReaderOrig TermVectorsReader
	pointsReader          PointsReader
	cfsReader             CompoundDirectory
	segment               string

	// fieldinfos for this core: means gen=-1. this is the exact fieldinfos these codec components saw at write.
	// in the case of DV updates, SR may hold a newer version.
	coreFieldInfos *FieldInfos

	// TODO: make a single thread local w/ a
	// Thingy class holding fieldsReader, termVectorsReader,
	// normsProducer

	fieldsReaderLocal   StoredFieldsReader
	termVectorsLocal    TermVectorsReader
	coreClosedListeners ClosedListener
}

func (r *SegmentCoreReaders) DecRef() error {
	if r.ref.Dec() == 0 {
		// TODO: close

	}
	return nil
}

func NewSegmentCoreReaders(dir store.Directory,
	si *SegmentCommitInfo, context store.IOContext) *SegmentCoreReaders {

	panic("")
}
