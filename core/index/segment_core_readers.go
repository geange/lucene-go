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
	si *SegmentCommitInfo, context *store.IOContext) (*SegmentCoreReaders, error) {

	codec := si.info.GetCodec()

	// confusing name: if (cfs) it's the cfsdir, otherwise it's the segment's directory.
	var cfsDir store.Directory
	//success := false

	r := &SegmentCoreReaders{}

	if si.info.GetUseCompoundFile() {
		reader, err := codec.CompoundFormat().GetCompoundReader(dir, si.info, context)
		if err != nil {
			return nil, err
		}
		cfsDir = reader
		r.cfsReader = reader
	} else {
		r.cfsReader = nil
		cfsDir = dir
	}

	var err error

	r.segment = si.info.name
	r.coreFieldInfos, err = codec.FieldInfosFormat().Read(cfsDir, si.info, "", context)
	if err != nil {
		return nil, err
	}

	segmentReadState := NewSegmentReadState(cfsDir, si.info, r.coreFieldInfos, context, "")
	format := codec.PostingsFormat()
	// Ask codec for its Fields
	r.fields, err = format.FieldsProducer(segmentReadState)
	if err != nil {
		return nil, err
	}

	// ask codec for its Norms:
	// TODO: since we don't write any norms file if there are no norms,
	// kinda jaky to assume the codec handles the case of no norms file at all gracefully?!

	if r.coreFieldInfos.HasNorms() {
		r.normsProducer, err = codec.NormsFormat().NormsProducer(segmentReadState)
		if err != nil {
			return nil, err
		}
		//assert normsProducer != null;
	} else {
		r.normsProducer = nil
	}

	r.fieldsReaderOrig, err = si.info.GetCodec().StoredFieldsFormat().FieldsReader(cfsDir, si.info, r.coreFieldInfos, context)
	if err != nil {
		return nil, err
	}
	r.fieldsReaderLocal = r.fieldsReaderOrig.Clone()

	if r.coreFieldInfos.HasVectors() { // open term vector files only as needed
		r.termVectorsReaderOrig, err = si.info.GetCodec().TermVectorsFormat().
			VectorsReader(cfsDir, si.info, r.coreFieldInfos, context)
		if err != nil {
			return nil, err
		}
		r.termVectorsLocal = r.termVectorsReaderOrig.Clone()
	} else {
		r.termVectorsReaderOrig = nil
	}

	if r.coreFieldInfos.HasPointValues() {
		r.pointsReader, err = codec.PointsFormat().FieldsReader(segmentReadState)
		if err != nil {
			return nil, err
		}
	} else {
		r.pointsReader = nil
	}

	return r, nil
}
