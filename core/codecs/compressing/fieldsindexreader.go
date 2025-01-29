package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ FieldsIndex = &FieldsIndexReader{}

const (
	VERSION_START   = 0
	VERSION_CURRENT = 0
)

type FieldsIndexReader struct {
	maxDoc                    int
	blockShift                int
	numChunks                 int
	docsMeta                  *packed.Meta
	startPointersMeta         *packed.Meta
	indexInput                store.IndexInput
	docsStartPointer          int64
	docsEndPointer            int64
	startPointersStartPointer int64
	startPointersEndPointer   int64
	docs                      *packed.DirectMonotonicReader
	startPointers             *packed.DirectMonotonicReader
	maxPointer                int64
}

func NewFieldsIndexReader(ctx context.Context, dir store.Directory, name, suffix, extension, codecName string,
	id []byte, metaIn store.IndexInput) (*FieldsIndexReader, error) {

	maxDoc, err := metaIn.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	blockShift, err := metaIn.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	numChunks, err := metaIn.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	docsStartPointer, err := metaIn.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}
	docsMeta, err := packed.LoadMeta(metaIn, int64(numChunks), int(blockShift))
	if err != nil {
		return nil, err
	}
	pointer, err := metaIn.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}
	docsEndPointer := pointer
	startPointersStartPointer := pointer
	startPointersMeta, err := packed.LoadMeta(metaIn, int64(numChunks), int(blockShift))
	if err != nil {
		return nil, err
	}
	startPointersEndPointer, err := metaIn.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}
	maxPointer, err := metaIn.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}

	indexInput, err := dir.OpenInput(ctx, store.SegmentFileName(name, suffix, extension))
	if err != nil {
		return nil, err
	}
	if _, err := codecs.CheckIndexHeader(ctx, indexInput, codecName+"Idx",
		VERSION_START, VERSION_CURRENT, id, suffix); err != nil {

		return nil, err
	}

	if _, err := codecs.RetrieveChecksum(ctx, indexInput); err != nil {
		return nil, err
	}

	docsSlice, err := indexInput.RandomAccessSlice(int64(docsStartPointer), int64(docsEndPointer-docsStartPointer))
	if err != nil {
		return nil, err
	}

	startPointersSlice, err := indexInput.RandomAccessSlice(int64(startPointersStartPointer),
		int64(startPointersEndPointer-startPointersStartPointer))
	if err != nil {
		return nil, err
	}

	docs, err := packed.DirectMonotonicReaderGetInstance(docsMeta, docsSlice)
	if err != nil {
		return nil, err
	}

	startPointers, err := packed.DirectMonotonicReaderGetInstance(startPointersMeta, startPointersSlice)
	if err != nil {
		return nil, err
	}

	reader := &FieldsIndexReader{
		maxDoc:                    int(maxDoc),
		blockShift:                int(blockShift),
		numChunks:                 int(numChunks),
		docsMeta:                  docsMeta,
		startPointersMeta:         startPointersMeta,
		indexInput:                indexInput,
		docsStartPointer:          int64(docsStartPointer),
		docsEndPointer:            int64(docsEndPointer),
		startPointersStartPointer: int64(startPointersStartPointer),
		startPointersEndPointer:   int64(startPointersEndPointer),
		docs:                      docs,
		startPointers:             startPointers,
		maxPointer:                int64(maxPointer),
	}

	return reader, nil

	//     indexInput = dir.openInput(IndexFileNames.segmentFileName(name, suffix, extension), IOContext.READ);
	//    boolean success = false;
	//    try {
	//      CodecUtil.checkIndexHeader(indexInput, codecName + "Idx", VERSION_START, VERSION_CURRENT, id, suffix);
	//      CodecUtil.retrieveChecksum(indexInput);
	//      success = true;
	//    } finally {
	//      if (success == false) {
	//        indexInput.close();
	//      }
	//    }
	//    final RandomAccessInput docsSlice = indexInput.randomAccessSlice(docsStartPointer, docsEndPointer - docsStartPointer);
	//    final RandomAccessInput startPointersSlice = indexInput.randomAccessSlice(startPointersStartPointer, startPointersEndPointer - startPointersStartPointer);
	//    docs = DirectMonotonicReader.getInstance(docsMeta, docsSlice);
	//    startPointers = DirectMonotonicReader.getInstance(startPointersMeta, startPointersSlice);
}

func (f *FieldsIndexReader) Close() error {
	return f.indexInput.Close()
}

func (f *FieldsIndexReader) GetStartPointer(docID int) int64 {
	blockIndex, err := f.docs.BinarySearch(0, int64(f.numChunks), int64(docID))
	if err != nil {
		return -1
	}
	if blockIndex < 0 {
		blockIndex = -2 - blockIndex
	}
	pointer, _ := f.startPointers.Get(int(blockIndex))
	return pointer
}

func (f *FieldsIndexReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldsIndexReader) Clone() FieldsIndex {
	//TODO implement me
	panic("implement me")
}
