package lucene80

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/packed"
	"github.com/pierrec/lz4/v4"
)

var _ index.DocValuesConsumer = &DocValuesConsumer{}

type DocValuesConsumer struct {
	mode *DocValuesConsumerMode

	data, meta      store.IndexOutput
	maxDoc          int
	state           *index.SegmentWriteState
	termsDictBuffer []byte
}

type DocValuesConsumerMode byte

const (
	BEST_SPEED = DocValuesConsumerMode(iota)
	BEST_COMPRESSION
)

func NewDocValuesConsumer(ctx context.Context, state *index.SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string, mode DocValuesConsumerMode) (*DocValuesConsumer, error) {
	panic("implement me")
}

func (d *DocValuesConsumer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddBinaryField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddSortedField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddSortedNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddSortedSetField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) NewCompressedBinaryBlockWriter(ctx context.Context) (*CompressedBinaryBlockWriter, error) {
	tempBinaryOffsets, err := d.state.Directory.CreateTempOutput(ctx, d.state.SegmentInfo.Name(), "binary_pointers")
	if err != nil {
		return nil, err
	}
	if err := codecs.WriteHeader(ctx, tempBinaryOffsets, DV_META_CODEC+"FilePointers", DV_VERSION_CURRENT); err != nil {
		return nil, err
	}
	blockAddressesStart := d.data.GetFilePointer()
	writer := lz4.NewWriter(d.data)

	return &CompressedBinaryBlockWriter{
		consumer:            d,
		lz4writer:           writer,
		tempBinaryOffsets:   tempBinaryOffsets,
		blockAddressesStart: blockAddressesStart,
	}, nil
}

type CompressedBinaryBlockWriter struct {
	consumer  *DocValuesConsumer
	lz4writer *lz4.Writer
	//uncompressedBlockLength    int
	maxUncompressedBlockLength int
	numDocsInCurrentBlock      int
	docLengths                 []int
	block                      []byte
	totalChunks                int
	maxPointer                 int64
	blockAddressesStart        int64
	tempBinaryOffsets          store.IndexOutput
}

func (w *CompressedBinaryBlockWriter) addDoc(ctx context.Context, v []byte) error {
	w.docLengths = append(w.docLengths, len(v))
	w.block = append(w.block, v...)

	w.numDocsInCurrentBlock++
	if w.numDocsInCurrentBlock == DV_BINARY_DOCS_PER_COMPRESSED_BLOCK {
		if err := w.flushData(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (w *CompressedBinaryBlockWriter) flushData(ctx context.Context) error {
	if w.numDocsInCurrentBlock <= 0 {
		return nil
	}

	data := w.consumer.data

	// Write offset to this block to temporary offsets file
	w.totalChunks++
	thisBlockStartPointer := data.GetFilePointer()

	// Optimisation - check if all lengths are same
	allLengthsSame := true
	for i := 1; i < DV_BINARY_DOCS_PER_COMPRESSED_BLOCK; i++ {
		if w.docLengths[i] != w.docLengths[i-1] {
			allLengthsSame = false
			break
		}
	}
	if allLengthsSame {
		// Only write one value shifted. Steal a bit to indicate all other lengths are the same
		onlyOneLength := (w.docLengths[0] << 1) | 1
		if err := data.WriteUvarint(ctx, uint64(onlyOneLength)); err != nil {
			return err
		}
	} else {
		for i := 0; i < DV_BINARY_DOCS_PER_COMPRESSED_BLOCK; i++ {
			if i == 0 {
				// Write first value shifted and steal a bit to indicate other lengths are to follow
				multipleLengths := w.docLengths[0] << 1
				if err := data.WriteUvarint(ctx, uint64(multipleLengths)); err != nil {
					return err
				}
				continue
			}
			if err := data.WriteUvarint(ctx, uint64(w.docLengths[i])); err != nil {
				return err
			}
		}
	}
	w.maxUncompressedBlockLength = max(w.maxUncompressedBlockLength, len(w.block))
	if _, err := w.lz4writer.Write(w.block); err != nil {
		return err
	}
	if err := w.lz4writer.Flush(); err != nil {
		return err
	}
	w.block = w.block[:0]
	w.docLengths = w.docLengths[:0]
	w.numDocsInCurrentBlock = 0
	// Ensure initialized with zeroes because full array is always written
	w.maxPointer = data.GetFilePointer()
	if err := w.tempBinaryOffsets.WriteUvarint(ctx, uint64(w.maxPointer-thisBlockStartPointer)); err != nil {
		return err
	}
	return nil
}

func (w *CompressedBinaryBlockWriter) writeMetaData(ctx context.Context) error {
	if w.totalChunks == 0 {
		return nil
	}

	data := w.consumer.data
	meta := w.consumer.meta

	startDMW := data.GetFilePointer()
	if err := meta.WriteUint64(ctx, uint64(startDMW)); err != nil {
		return err
	}

	if err := meta.WriteUvarint(ctx, uint64(w.totalChunks)); err != nil {
		return err
	}
	if err := meta.WriteUvarint(ctx, DV_BINARY_BLOCK_SHIFT); err != nil {
		return err
	}
	if err := meta.WriteUvarint(ctx, uint64(w.maxUncompressedBlockLength)); err != nil {
		return err
	}
	if err := meta.WriteUvarint(ctx, DV_DIRECT_MONOTONIC_BLOCK_SHIFT); err != nil {
		return err
	}

	if err := codecs.WriteFooter(ctx, w.tempBinaryOffsets); err != nil {
		return err
	}
	if err := util.Close(w.tempBinaryOffsets); err != nil {
		return err
	}

	state := w.consumer.state
	//write the compressed block offsets info to the meta file by reading from temp file
	filePointersIn, err := store.OpenChecksumInput(ctx, state.Directory, w.tempBinaryOffsets.GetName())
	if err != nil {
		return err
	}
	if _, err := codecs.CheckHeader(ctx, filePointersIn, DV_META_CODEC+"FilePointers", DV_VERSION_CURRENT, DV_VERSION_CURRENT); err != nil {
		return err
	}

	filePointers, err := packed.DirectMonotonicWriterGetInstance(meta, data, int64(w.totalChunks), DV_DIRECT_MONOTONIC_BLOCK_SHIFT)
	if err != nil {
		return err
	}

	fp := w.blockAddressesStart
	for i := 0; i < w.totalChunks; i++ {
		if err := filePointers.Add(fp); err != nil {
			return err
		}

		n, err := filePointersIn.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		fp += int64(n)
	}
	if w.maxPointer < fp {
		return errors.New("max pointer out of bounds")
	}

	if err := filePointers.Finish(); err != nil {
		return err
	}
	if _, err := codecs.CheckFooter(ctx, filePointersIn); err != nil {
		return err
	}

	// Write the length of the DMW block in the data
	if err := meta.WriteUint64(ctx, uint64(data.GetFilePointer()-startDMW)); err != nil {
		return err
	}
	return nil
}
