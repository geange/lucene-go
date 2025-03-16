package lucene86

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bkd"
)

var _ index.PointsWriter = &PointsWriter{}

type PointsWriter struct {
	*coreIndex.BasePointsWriter

	metaOut             store.IndexOutput
	indexOut            store.IndexOutput
	dataOut             store.IndexOutput
	writeState          *index.SegmentWriteState
	maxPointsInLeafNode int
	maxMBSortInHeap     float64
	finished            bool
}

type PointsWriterConfig struct {
	writeState          *index.SegmentWriteState
	maxPointsInLeafNode int
	maxMBSortInHeap     float64
}

func NewPointsWriterConfig(writeState *index.SegmentWriteState) *PointsWriterConfig {
	return &PointsWriterConfig{
		writeState:          writeState,
		maxPointsInLeafNode: bkd.DEFAULT_MAX_POINTS_IN_LEAF_NODE,
		maxMBSortInHeap:     bkd.DEFAULT_MAX_MB_SORT_IN_HEAP,
	}
}

func (p *PointsWriterConfig) WithMaxPointsInLeafNode(maxPointsInLeafNode int) *PointsWriterConfig {
	p.maxPointsInLeafNode = maxPointsInLeafNode
	return p
}

func (p *PointsWriterConfig) WithMaxMBSortInHeap(maxMBSortInHeap float64) *PointsWriterConfig {
	p.maxMBSortInHeap = maxMBSortInHeap
	return p
}

func NewPointsWriter(ctx context.Context, conf *PointsWriterConfig) (*PointsWriter, error) {
	writer := &PointsWriter{
		writeState:          conf.writeState,
		maxPointsInLeafNode: conf.maxPointsInLeafNode,
		maxMBSortInHeap:     conf.maxMBSortInHeap,
	}
	writer.BasePointsWriter = &coreIndex.BasePointsWriter{
		WriteField: writer.WriteField,
		Finish:     writer.Finish,
	}
	writeState := conf.writeState
	dataFileName := store.SegmentFileName(conf.writeState.SegmentInfo.Name(),
		conf.writeState.SegmentSuffix, POINT_DATA_EXTENSION)

	dataOut, err := writeState.Directory.CreateOutput(ctx, dataFileName)
	if err != nil {
		return nil, err
	}
	writer.dataOut = dataOut

	if err := codecs.WriteIndexHeader(ctx, dataOut, POINT_DATA_CODEC_NAME, POINT_VERSION_CURRENT,
		writeState.SegmentInfo.GetID(), writeState.SegmentSuffix); err != nil {
		return nil, err
	}

	metaFileName := store.SegmentFileName(writeState.SegmentInfo.Name(), writeState.SegmentSuffix,
		POINT_META_EXTENSION)
	metaOut, err := writeState.Directory.CreateOutput(ctx, metaFileName)
	if err != nil {
		return nil, err
	}
	writer.metaOut = metaOut

	if err := codecs.WriteIndexHeader(ctx, metaOut, POINT_META_CODEC_NAME, POINT_VERSION_CURRENT,
		writeState.SegmentInfo.GetID(), writeState.SegmentSuffix); err != nil {
		return nil, err
	}

	indexFileName := store.SegmentFileName(writeState.SegmentInfo.Name(),
		writeState.SegmentSuffix, POINT_INDEX_EXTENSION)
	indexOut, err := writeState.Directory.CreateOutput(ctx, indexFileName)
	writer.indexOut = indexOut

	if err := codecs.WriteIndexHeader(ctx, indexOut, POINT_INDEX_CODEC_NAME, POINT_VERSION_CURRENT,
		writeState.SegmentInfo.GetID(), writeState.SegmentSuffix); err != nil {
		return nil, err
	}
	return writer, nil
}

func (p *PointsWriter) Close() error {
	if err := p.metaOut.Close(); err != nil {
		return err
	}
	if err := p.indexOut.Close(); err != nil {
		return err
	}
	if err := p.dataOut.Close(); err != nil {
		return err
	}
	return nil
}

func (p *PointsWriter) WriteField(ctx context.Context, fieldInfo *document.FieldInfo, reader index.PointsReader) error {
	values, err := reader.GetValues(ctx, fieldInfo.Name())
	if err != nil {
		return err
	}
	config, err := bkd.NewConfig(fieldInfo.GetPointDimensionCount(),
		fieldInfo.GetPointIndexDimensionCount(),
		fieldInfo.GetPointNumBytes(),
		p.maxPointsInLeafNode)
	if err != nil {
		return err
	}

	maxDoc, err := p.writeState.SegmentInfo.MaxDoc()
	if err != nil {
		return err
	}
	writer, err := bkd.NewWriter(maxDoc, p.writeState.Directory, p.writeState.SegmentInfo.Name(), config,
		p.maxMBSortInHeap, values.Size())
	if err != nil {
		return err
	}

	mutablePointValues, ok := values.(types.MutablePointValues)
	if ok {
		finalizer, err := writer.WriteField(ctx, p.metaOut, p.indexOut, p.dataOut, fieldInfo.Name(), mutablePointValues)
		if err != nil {
			return err
		}
		if err := finalizer(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *PointsWriter) Finish() error {
	if p.finished {
		return errors.New("already finished")
	}
	ctx := context.Background()
	p.finished = true
	n := int32(-1)
	if err := p.metaOut.WriteUint32(ctx, uint32(n)); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, p.indexOut); err != nil {
		return err
	}
	if err := p.metaOut.WriteUint64(ctx, uint64(p.indexOut.GetFilePointer())); err != nil {
		return err
	}
	if err := p.metaOut.WriteUint64(ctx, uint64(p.dataOut.GetFilePointer())); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, p.metaOut); err != nil {
		return err
	}

	return nil
}

//func (p *PointsWriter) Merge(mergeState *index.MergeState) error {
//	// If indexSort is activated and some of the leaves are not sorted the next test will catch that and the non-optimized merge will run.
//	// If the readers are all sorted then it's safe to perform a bulk merge of the points.
//	for _, reader := range mergeState.PointsReaders {
//		_, ok := reader.(*PointsReader)
//		if !ok {
//			return p.BasePointsWriter.Merge(mergeState)
//		}
//	}
//
//	for _, reader := range mergeState.PointsReaders {
//		if reader != nil {
//			if err := reader.CheckIntegrity(); err != nil {
//				return err
//			}
//		}
//	}
//
//	ctx := context.Background()
//
//	for _, fieldInfo := range mergeState.MergeFieldInfos.List() {
//		if fieldInfo.GetPointDimensionCount() == 0 {
//			continue
//		}
//
//		if fieldInfo.GetPointDimensionCount() == 1 {
//
//			// Worst case total maximum size (if none of the points are deleted):
//			totMaxSize := 0
//			for i, reader := range mergeState.PointsReaders {
//				if reader == nil {
//					continue
//				}
//				readerFieldInfos := mergeState.FieldInfos[i]
//				readerFieldInfo := readerFieldInfos.FieldInfo(fieldInfo.Name())
//
//				if readerFieldInfo != nil && readerFieldInfo.GetPointDimensionCount() > 0 {
//					values, err := reader.GetValues(ctx, fieldInfo.Name())
//					if err != nil {
//						return err
//					}
//					totMaxSize += values.Size()
//				}
//			}
//
//			config, err := bkd.NewConfig(fieldInfo.GetPointDimensionCount(),
//				fieldInfo.GetPointIndexDimensionCount(), fieldInfo.GetPointNumBytes(),
//				p.maxPointsInLeafNode)
//			if err != nil {
//				return err
//			}
//
//			// Optimize the 1D case to use BKDWriter.merge, which does a single merge sort of the
//			// already sorted incoming segments, instead of trying to sort all points again as if
//			// we were simply reindexing them:
//			maxDoc, err := p.writeState.SegmentInfo.MaxDoc()
//			if err != nil {
//				return err
//			}
//			writer, err := bkd.NewWriter(maxDoc, p.writeState.Directory, p.writeState.SegmentInfo.Name(),
//				config, p.maxMBSortInHeap, totMaxSize)
//			if err != nil {
//				return err
//			}
//
//			bkdReaders := make([]*bkd.Reader, 0)
//			docMaps := make([]index.MergeStateDocMap, 0)
//
//			for i, reader := range mergeState.PointsReaders {
//				if reader == nil {
//					continue
//				}
//				pointReader, ok := reader.(*PointsReader)
//				if !ok {
//					// TODO:
//				}
//
//				// NOTE: we cannot just use the merged fieldInfo.number (instead of resolving to this
//				// reader's FieldInfo as we do below) because field numbers can easily be different
//				// when addIndexes(Directory...) copies over segments from another index:
//
//				readerFieldInfos := mergeState.FieldInfos[i]
//				readerFieldInfo := readerFieldInfos.FieldInfo(fieldInfo.Name())
//
//				if readerFieldInfo != nil && readerFieldInfo.GetPointDimensionCount() > 0 {
//					bkdReader, exist := pointReader.readers[readerFieldInfo.Number()]
//					if exist {
//						bkdReaders = append(bkdReaders, bkdReader)
//						docMaps = append(docMaps, mergeState.DocMaps[i])
//					}
//				}
//
//			}
//
//			finalizer, err := writer.Merge(ctx, p.metaOut, p.indexOut, p.dataOut, docMaps, bkdReaders)
//			if err != nil {
//				return err
//			}
//			if err := finalizer(ctx); err != nil {
//				return err
//			}
//		} else {
//			if err := p.MergeOneField(ctx, mergeState, fieldInfo); err != nil {
//				return err
//			}
//		}
//	}
//
//	return p.Finish()
//}
