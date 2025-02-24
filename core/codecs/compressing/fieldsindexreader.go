package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ FieldsIndex = &FieldsIndexReader{}

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
		FIELDS_VERSION_START, FIELDS_VERSION_CURRENT, id, suffix); err != nil {

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
}

func newFieldsIndexReader(other *FieldsIndexReader) (*FieldsIndexReader, error) {
	maxDoc := other.maxDoc
	numChunks := other.numChunks
	blockShift := other.blockShift
	docsMeta := other.docsMeta
	startPointersMeta := other.startPointersMeta
	indexInput := other.indexInput.Clone().(store.IndexInput)
	docsStartPointer := other.docsStartPointer
	docsEndPointer := other.docsEndPointer
	startPointersStartPointer := other.startPointersStartPointer
	startPointersEndPointer := other.startPointersEndPointer
	maxPointer := other.maxPointer
	docsSlice, err := indexInput.RandomAccessSlice(docsStartPointer,
		docsEndPointer-docsStartPointer)
	if err != nil {
		return nil, err
	}
	startPointersSlice, err := indexInput.RandomAccessSlice(startPointersStartPointer,
		startPointersEndPointer-startPointersStartPointer)
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

	return &FieldsIndexReader{
		maxDoc:                    maxDoc,
		blockShift:                blockShift,
		numChunks:                 numChunks,
		docsMeta:                  docsMeta,
		startPointersMeta:         startPointersMeta,
		indexInput:                indexInput,
		docsStartPointer:          docsStartPointer,
		docsEndPointer:            docsEndPointer,
		startPointersStartPointer: startPointersStartPointer,
		startPointersEndPointer:   startPointersEndPointer,
		docs:                      docs,
		startPointers:             startPointers,
		maxPointer:                maxPointer,
	}, nil
}

func (f *FieldsIndexReader) Close() error {
	return f.indexInput.Close()
}

func (f *FieldsIndexReader) GetStartPointer(docID int) (int64, error) {
	blockIndex, err := f.docs.BinarySearch(0, int64(f.numChunks), int64(docID))
	if err != nil {
		return 0, err
	}
	if blockIndex < 0 {
		blockIndex = -2 - blockIndex
	}
	pointer, err := f.startPointers.Get(int(blockIndex))
	if err != nil {
		return 0, err
	}
	return pointer, nil
}

func (f *FieldsIndexReader) CheckIntegrity() error {
	if _, err := codecs.ChecksumEntireFile(context.Background(), f.indexInput); err != nil {
		return err
	}
	return nil
}

func (f *FieldsIndexReader) Clone() (FieldsIndex, error) {
	return newFieldsIndexReader(f)
}

func (f *FieldsIndexReader) GetMaxPointer() int {
	return int(f.maxPointer)
}
