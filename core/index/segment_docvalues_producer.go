package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
)

var _ DocValuesProducer = &SegmentDocValuesProducer{}

type SegmentDocValuesProducer struct {
	dvProducersByField map[string]DocValuesProducer
	dvProducers        []DocValuesProducer
	dvGens             []int64
}

func NewSegmentDocValuesProducer(si *SegmentCommitInfo, dir store.Directory,
	coreInfos, allInfos *FieldInfos, segDocValues *SegmentDocValues) (*SegmentDocValuesProducer, error) {

	p := &SegmentDocValuesProducer{
		dvProducersByField: map[string]DocValuesProducer{},
		dvProducers:        []DocValuesProducer{},
		dvGens:             []int64{},
	}

	var baseProducer DocValuesProducer
	for _, fi := range allInfos.values {
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

func (s *SegmentDocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetNumeric(field *document.FieldInfo) (NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetBinary(field *document.FieldInfo) (BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetSorted(field *document.FieldInfo) (SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetSortedNumeric(field *document.FieldInfo) (SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) GetSortedSet(field *document.FieldInfo) (SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentDocValuesProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}
