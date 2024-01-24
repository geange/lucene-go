package index

import (
	"errors"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/packed"
)

// MergeState Holds common state used during segment merging.
type MergeState struct {
	// Maps document IDs from old segments to document IDs in the new segment
	DocMaps []MergeStateDocMap

	// SegmentInfo of the newly merged segment.
	SegmentInfo *SegmentInfo

	// FieldInfos of the newly merged segment.
	MergeFieldInfos *FieldInfos

	// Stored field producers being merged
	StoredFieldsReaders []StoredFieldsReader

	// Term vector producers being merged
	TermVectorsReaders []TermVectorsReader

	// Norms producers being merged
	NormsProducers []NormsProducer

	// DocValues producers being merged
	DocValuesProducers []DocValuesProducer

	// FieldInfos being merged
	FieldInfos []*FieldInfos

	// Live docs for each reader
	LiveDocs []util.Bits

	// Postings to merge
	FieldsProducers []FieldsProducer

	// Point readers to merge
	PointsReaders []PointsReader

	// Max docs per reader
	MaxDocs []int

	// InfoStream for debugging messages.

	// Indicates if the index needs to be sorted
	NeedsIndexSort bool
}

func NewMergeState(readers []CodecReader, segmentInfo *SegmentInfo) (*MergeState, error) {
	if err := verifyIndexSort(readers, segmentInfo); err != nil {
		return nil, err
	}

	numReaders := len(readers)

	state := MergeState{
		StoredFieldsReaders: make([]StoredFieldsReader, numReaders),
		TermVectorsReaders:  make([]TermVectorsReader, numReaders),
		NormsProducers:      make([]NormsProducer, numReaders),
		DocValuesProducers:  make([]DocValuesProducer, numReaders),
		FieldInfos:          make([]*FieldInfos, numReaders),
		LiveDocs:            make([]util.Bits, numReaders),
		FieldsProducers:     make([]FieldsProducer, numReaders),
		PointsReaders:       make([]PointsReader, numReaders),
		MaxDocs:             make([]int, numReaders),
	}

	numDocs := 0
	for i, reader := range readers {
		state.MaxDocs[i] = reader.MaxDoc()
		state.LiveDocs[i] = reader.GetLiveDocs()
		state.FieldInfos[i] = reader.GetFieldInfos()

		state.NormsProducers[i] = reader.GetNormsReader()
		if state.NormsProducers[i] != nil {
			state.NormsProducers[i] = state.NormsProducers[i].GetMergeInstance()
		}

		state.DocValuesProducers[i] = reader.GetDocValuesReader()
		if state.DocValuesProducers[i] != nil {
			state.DocValuesProducers[i] = state.DocValuesProducers[i].GetMergeInstance()
		}

		state.StoredFieldsReaders[i] = reader.GetFieldsReader()
		if state.StoredFieldsReaders[i] != nil {
			state.StoredFieldsReaders[i] = state.StoredFieldsReaders[i].GetMergeInstance()
		}

		state.TermVectorsReaders[i] = reader.GetTermVectorsReader()
		if state.TermVectorsReaders[i] != nil {
			state.TermVectorsReaders[i] = state.TermVectorsReaders[i].GetMergeInstance()
		}

		state.FieldsProducers[i] = reader.GetPostingsReader().GetMergeInstance()
		state.PointsReaders[i] = reader.GetPointsReader()
		if state.PointsReaders[i] != nil {
			state.PointsReaders[i] = state.PointsReaders[i].GetMergeInstance()
		}
		numDocs += reader.NumDocs()
	}

	if err := segmentInfo.SetMaxDoc(numDocs); err != nil {
		return nil, err
	}
	state.SegmentInfo = segmentInfo
	state.DocMaps = buildDocMaps(readers, segmentInfo.GetIndexSort())
	return &state, nil
}

func verifyIndexSort(readers []CodecReader, segmentInfo *SegmentInfo) error {
	indexSort := segmentInfo.GetIndexSort()
	if indexSort == nil {
		return nil
	}

	for _, leaf := range readers {
		segmentSort := leaf.GetMetaData().GetSort()
		if segmentSort == nil || isCongruentSort(indexSort, segmentSort) == false {
			return errors.New("index sort mismatch")
		}
	}
	return nil
}

func buildDocMaps(readers []CodecReader, indexSort *Sort) []MergeStateDocMap {
	if indexSort == nil {
		// no index sort ... we only must map around deletions, and rebase to the merged segment's docID space
		return buildDeletionDocMaps(readers)
	}

	// do a merge sort of the incoming leaves:
	//to := time.Now().UnixNano()
	panic("")
}

func buildDeletionDocMaps(readers []CodecReader) []MergeStateDocMap {
	docMaps := make([]MergeStateDocMap, 0, len(readers))
	var totalDocs int

	for _, reader := range readers {
		liveDocs := reader.GetLiveDocs()

		var delDocMap *packed.LongValues
		if liveDocs != nil {
			delDocMap = removeDeletes(reader.MaxDoc(), liveDocs)
		} else {
			delDocMap = nil
		}

		docBase := totalDocs

		docMaps = append(docMaps, MergeStateDocMap{func(docID int) int {
			if liveDocs == nil {
				return docBase + docID
			} else if liveDocs.Test(uint(docID)) {
				return docBase + int(delDocMap.Get(int64(docID)))
			} else {
				return -1
			}
		}})

		totalDocs += reader.NumDocs()
	}
	return docMaps
}

func removeDeletes(maxDoc int, liveDocs util.Bits) *packed.LongValues {
	// TODO: fix it
	panic("")
	//docMapBuilder := packed.NewLongValuesBuilder()
	//del := 0
	//
	//for i := 0; i < maxDoc; i++ {
	//	docMapBuilder.Add(int64(i - del))
	//	if !liveDocs.Test(uint(i)) {
	//		del++
	//	}
	//}
	//return docMapBuilder.Build()
}

type MergeStateDocMap struct {
	Get func(docId int) int
}
