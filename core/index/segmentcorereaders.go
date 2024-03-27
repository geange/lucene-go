package index

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"io"
	"sync/atomic"
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

func NewSegmentCoreReaders(ctx context.Context, dir store.Directory, si *SegmentCommitInfo, ioContext *store.IOContext) (*SegmentCoreReaders, error) {

	codec := si.info.GetCodec()

	// confusing name: if (cfs) it's the cfsdir, otherwise it's the segment's directory.
	var cfsDir store.Directory

	r := &SegmentCoreReaders{}

	if si.info.GetUseCompoundFile() {
		reader, err := codec.CompoundFormat().GetCompoundReader(ctx, dir, si.info, ioContext)
		if err != nil {
			return nil, err
		}
		cfsDir = reader
		r.cfsReader = reader
	} else {
		r.cfsReader = nil
		cfsDir = dir
	}

	r.segment = si.info.Name()
	coreFieldInfos, err := codec.FieldInfosFormat().Read(ctx, cfsDir, si.info, "", ioContext)
	if err != nil {
		return nil, err
	}
	r.coreFieldInfos = coreFieldInfos

	segmentReadState := NewSegmentReadState(cfsDir, si.info, r.coreFieldInfos, ioContext, "")

	// Ask codec for its Fields
	fields, err := codec.PostingsFormat().FieldsProducer(ctx, segmentReadState)
	if err != nil {
		return nil, err
	}
	r.fields = fields

	// ask codec for its Norms:
	// TODO: since we don't write any norms file if there are no norms,
	// kinda jaky to assume the codec handles the case of no norms file at all gracefully?!

	if r.coreFieldInfos.HasNorms() {
		normsProducer, err := codec.NormsFormat().NormsProducer(ctx, segmentReadState)
		if err != nil {
			return nil, err
		}
		r.normsProducer = normsProducer
		//assert normsProducer != null;
	} else {
		r.normsProducer = nil
	}

	fieldsReaderOrig, err := si.info.GetCodec().StoredFieldsFormat().FieldsReader(ctx, cfsDir, si.info, r.coreFieldInfos, ioContext)
	if err != nil {
		return nil, err
	}
	r.fieldsReaderOrig = fieldsReaderOrig
	r.fieldsReaderLocal = fieldsReaderOrig.Clone(ctx)

	if r.coreFieldInfos.HasVectors() { // open term vector files only as needed
		termVectorsReaderOrig, err := si.info.GetCodec().TermVectorsFormat().
			VectorsReader(nil, cfsDir, si.info, r.coreFieldInfos, ioContext)
		if err != nil {
			return nil, err
		}
		r.termVectorsReaderOrig = termVectorsReaderOrig
		r.termVectorsLocal = termVectorsReaderOrig.Clone(ctx)
	} else {
		r.termVectorsReaderOrig = nil
	}

	if r.coreFieldInfos.HasPointValues() {
		r.pointsReader, err = codec.PointsFormat().FieldsReader(ctx, segmentReadState)
		if err != nil {
			return nil, err
		}
	} else {
		r.pointsReader = nil
	}

	return r, nil
}

func (s *SegmentCoreReaders) incRef() error {
	if s.ref.Load() < 0 {
		return errors.New("segmentCoreReaders is already closed")
	}
	s.ref.Add(1)
	return nil
}

func (s *SegmentCoreReaders) decRef() error {
	if s.ref.Add(-1) == 0 {

		closers := []io.Closer{
			s.termVectorsLocal,
			s.fieldsReaderLocal,
			s.fields,
			s.termVectorsReaderOrig,
			s.fieldsReaderOrig,
			s.cfsReader,
			s.normsProducer,
			s.pointsReader,
		}

		if err := closeAll(closers...); err != nil {
			return err
		}
	}
	return nil
}

func closeAll(objects ...io.Closer) error {
	for _, object := range objects {
		if err := object.Close(); err != nil {
			return err
		}
	}
	return nil
}
