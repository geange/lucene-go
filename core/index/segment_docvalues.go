package index

import (
	"strconv"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

// SegmentDocValues
// Manages the DocValuesProducer held by SegmentReader and keeps track of their reference counting.
type SegmentDocValues struct {
	genDVProducers map[int64]*util.RefCount[DocValuesProducer]
}

func (s *SegmentDocValues) GetDocValuesProducer(gen int64,
	si *SegmentCommitInfo, dir store.Directory, infos *FieldInfos) (DocValuesProducer, error) {

	dvp, ok := s.genDVProducers[gen]
	if !ok {
		var err error
		dvp, err = s.newDocValuesProducer(si, dir, gen, infos)
		if err != nil {
			return nil, err
		}
		//assert dvp != null;
		s.genDVProducers[gen] = dvp
	} else {
		dvp.IncRef()
	}
	return dvp.Get(), nil
}

func (s *SegmentDocValues) newDocValuesProducer(si *SegmentCommitInfo,
	dir store.Directory, gen int64, infos *FieldInfos) (*util.RefCount[DocValuesProducer], error) {

	dvDir := dir
	segmentSuffix := ""
	if gen != -1 {
		dvDir = si.info.dir // gen'd files are written outside CFS, so use SegInfo directory
		segmentSuffix = strconv.FormatInt(gen, 36)
	}

	// set SegmentReadState to list only the fields that are relevant to that gen
	srs := NewSegmentReadState(dvDir, si.info, infos, nil, segmentSuffix)
	dvFormat := si.info.GetCodec().DocValuesFormat()

	producer, err := dvFormat.FieldsProducer(srs)
	if err != nil {
		return nil, err
	}
	return util.NewRefCount[DocValuesProducer](producer, func(r *util.RefCount[DocValuesProducer]) error {
		if err := r.Get().Close(); err != nil {
			return err
		}
		delete(s.genDVProducers, gen)
		return nil
	}), nil
}

func (s *SegmentDocValues) decRef(gens []int64) error {
	for _, gen := range gens {
		dvp, ok := s.genDVProducers[gen]
		if ok {
			if err := dvp.DecRef(); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewSegmentDocValues() *SegmentDocValues {
	return &SegmentDocValues{genDVProducers: map[int64]*util.RefCount[DocValuesProducer]{}}
}
