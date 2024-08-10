package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"strconv"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

// SegmentDocValues
// Manages the DocValuesProducer held by SegmentReader and keeps track of their reference counting.
type SegmentDocValues struct {
	genDVProducers map[int64]*util.RefCount[index.DocValuesProducer]
}

func (s *SegmentDocValues) GetDocValuesProducer(gen int64,
	si index.SegmentCommitInfo, dir store.Directory, infos index.FieldInfos) (index.DocValuesProducer, error) {

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

func (s *SegmentDocValues) newDocValuesProducer(si index.SegmentCommitInfo,
	dir store.Directory, gen int64, infos index.FieldInfos) (*util.RefCount[index.DocValuesProducer], error) {

	dvDir := dir
	segmentSuffix := ""
	if gen != -1 {
		dvDir = si.Info().Dir() // gen'd files are written outside CFS, so use SegInfo directory
		segmentSuffix = strconv.FormatInt(gen, 36)
	}

	// set SegmentReadState to list only the fields that are relevant to that gen
	srs := index.NewSegmentReadState(dvDir, si.Info(), infos, nil, segmentSuffix)
	dvFormat := si.Info().GetCodec().DocValuesFormat()

	producer, err := dvFormat.FieldsProducer(nil, srs)
	if err != nil {
		return nil, err
	}
	return util.NewRefCount[index.DocValuesProducer](producer, func(r *util.RefCount[index.DocValuesProducer]) error {
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
	return &SegmentDocValues{genDVProducers: map[int64]*util.RefCount[index.DocValuesProducer]{}}
}

var _ index.DocValuesProducer = &SegmentDocValuesProducer{}

type SegmentDocValuesProducer struct {
	dvProducersByField map[string]index.DocValuesProducer
	dvProducers        []index.DocValuesProducer
	dvGens             []int64
}

func NewSegmentDocValuesProducer(si index.SegmentCommitInfo, dir store.Directory,
	coreInfos, allInfos index.FieldInfos, segDocValues *SegmentDocValues) (*SegmentDocValuesProducer, error) {

	p := &SegmentDocValuesProducer{
		dvProducersByField: map[string]index.DocValuesProducer{},
		dvProducers:        []index.DocValuesProducer{},
		dvGens:             []int64{},
	}

	var baseProducer index.DocValuesProducer
	for _, fi := range allInfos.List() {
		if fi.GetDocValuesType() == document.DOC_VALUES_TYPE_NONE {
			continue
		}

		docValuesGen := fi.GetDocValuesGen()
		if docValuesGen == -1 {
			if baseProducer == nil {
				// the base producer gets the original fieldinfos it wrote
				var err error
				baseProducer, err = segDocValues.GetDocValuesProducer(docValuesGen, si, dir, coreInfos)
				if err != nil {
					return nil, err
				}
				p.dvGens = append(p.dvGens, docValuesGen)
				p.dvProducers = append(p.dvProducers, baseProducer)
			}
		} else {
			//assert !dvGens.contains(docValuesGen);
			// otherwise, producer sees only the one fieldinfo it wrote
			dvp, err := segDocValues.GetDocValuesProducer(docValuesGen, si, dir, NewFieldInfos([]*document.FieldInfo{fi}))
			if err != nil {
				return nil, err
			}
			p.dvGens = append(p.dvGens, docValuesGen)
			p.dvProducers = append(p.dvProducers, dvp)
			p.dvProducersByField[fi.Name()] = dvp
		}
	}
	return p, nil
}

func (s *SegmentDocValuesProducer) GetMergeInstance() index.DocValuesProducer {
	return s
}

func (s *SegmentDocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetNumeric(ctx context.Context, field *document.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetBinary(ctx context.Context, field *document.FieldInfo) (index.BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (index.SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetSortedNumeric(ctx context.Context, field *document.FieldInfo) (index.SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetSortedSet(ctx context.Context, field *document.FieldInfo) (index.SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}
