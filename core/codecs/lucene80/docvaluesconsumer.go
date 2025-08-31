package lucene80

import (
	"context"
	"errors"
	"io"
	"math"
	"math/big"
	"slices"

	"github.com/pierrec/lz4/v4"
	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ index.DocValuesConsumer = &DocValuesConsumer{}

type DocValuesConsumer struct {
	mode            *DocValuesConsumerMode
	data            store.IndexOutput
	meta            store.IndexOutput
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
	ctx := context.Background()

	eof := int32(-1)

	if d.meta != nil {
		// write EOF marker
		if err := d.meta.WriteUint32(ctx, uint32(eof)); err != nil {
			return err
		}
		// write checksum
		if err := codecs.WriteFooter(ctx, d.meta); err != nil {
			return err
		}
	}

	if d.data != nil {
		// write checksum
		if err := codecs.WriteFooter(ctx, d.data); err != nil {
			return err
		}
	}

	return util.Close(d.data, d.meta)
}

func (d *DocValuesConsumer) AddNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	if err := d.meta.WriteUint32(ctx, uint32(field.Number())); err != nil {
		return err
	}
	if err := d.meta.WriteByte(DV_NUMERIC); err != nil {
		return err
	}

	producer := &coreIndex.EmptyDocValuesProducer{
		FnGetSortedNumeric: func(ctx context.Context, field *document.FieldInfo) (index.SortedNumericDocValues, error) {
			values, err := valuesProducer.GetNumeric(ctx, field)
			if err != nil {
				return nil, err
			}
			return coreIndex.NewSingletonSortedNumericDocValues(values), nil
		},
	}

	_, err := d.writeValues(ctx, field, producer)
	if err != nil {
		return err
	}
	return nil
}

func (d *DocValuesConsumer) writeValues(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) ([]int64, error) {
	values, err := valuesProducer.GetSortedNumeric(ctx, field)
	if err != nil {
		return nil, err
	}
	numDocsWithValue := int64(0)
	minMax := newMinMaxTracker()
	blockMinMax := newMinMaxTracker()
	gcd := int64(0)
	uniqueValues := make(map[int64]struct{})

	for {
		_, err = values.NextDoc(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		count := values.DocValueCount()
		for i := 0; i < count; i++ {
			v, err := values.NextValue()
			if err != nil {
				return nil, err
			}

			if gcd != 1 {
				if v < math.MinInt64/2 || v > math.MaxInt64/2 {
					// in that case v - minValue might overflow and make the GCD computation return
					// wrong results. Since these extreme values are unlikely, we just discard
					// GCD computation for them
					gcd = 1
				} else if minMax.numValues != 0 { // minValue needs to be set first

					gcd = GCD(gcd, v-minMax.min)
				}
			}

			minMax.update(v)
			blockMinMax.update(v)
			if blockMinMax.numValues == DV_NUMERIC_BLOCK_SIZE {
				blockMinMax.nextBlock()
			}

			if uniqueValues != nil {
				uniqueValues[v] = struct{}{}

				if len(uniqueValues) > 256 {
					uniqueValues = nil
				}
			}

		}

		numDocsWithValue++
	}

	minMax.finish()
	blockMinMax.finish()

	numValues := minMax.numValues
	minV := minMax.min
	maxV := minMax.max

	if numDocsWithValue == 0 { // meta[-2, 0]: No documents with values
		store.WriteInt64(ctx, d.meta, -2) // docsWithFieldOffset
		d.meta.WriteUint64(ctx, 0)        // docsWithFieldLength
		d.meta.WriteUint16(ctx, -1)       // jumpTableEntryCount
		d.meta.WriteByte(-1)              // denseRankPower
	} else if numDocsWithValue == int64(d.maxDoc) { // meta[-1, 0]: All documents has values
		d.meta.WriteUint64(ctx, -1) // docsWithFieldOffset
		d.meta.WriteUint64(ctx, 0)  // docsWithFieldLength
		d.meta.WriteUint16(ctx, -1) // jumpTableEntryCount
		d.meta.WriteByte(-1)        // denseRankPower
	} else { // meta[data.offset, data.length]: IndexedDISI structure for documents with values
		offset := d.data.GetFilePointer()
		d.meta.WriteUint64(ctx, uint64(offset)) // docsWithFieldOffset
		values, err = valuesProducer.GetSortedNumeric(ctx, field)
		if err != nil {
			return nil, err
		}

		jumpTableEntryCount, err := WriteBitSet(ctx, values, d.data, DEFAULT_DENSE_RANK_POWER)
		if err != nil {
			return nil, err
		}
		d.meta.WriteUint64(ctx, uint64(d.data.GetFilePointer()-offset)) // docsWithFieldLength
		d.meta.WriteUint16(ctx, jumpTableEntryCount)
		d.meta.WriteByte(DEFAULT_DENSE_RANK_POWER)
	}

	d.meta.WriteUint64(ctx, uint64(numValues))
	var numBitsPerValue int
	doBlocks := false
	var encode map[int64]int
	if minV >= maxV { // meta[-1]: All values are 0
		numBitsPerValue = 0
		d.meta.WriteUint32(ctx, -1) // tablesize
	} else {
		if uniqueValues != nil && len(uniqueValues) > 1 &&
			packed.UnsignedBitsRequired(uint64(len(uniqueValues)-1)) <
				packed.UnsignedBitsRequired(uint64((maxV-minV)/gcd)) {
			numBitsPerValue = packed.UnsignedBitsRequired(uint64(len(uniqueValues) - 1))
			sortedUniqueValues := lo.Keys(uniqueValues)
			slices.Sort(sortedUniqueValues)
			d.meta.WriteUint32(ctx, uint32(len(sortedUniqueValues))) // tablesize
			for _, v := range sortedUniqueValues {
				d.meta.WriteUint64(ctx, uint64(v)) // table[] entry
			}
			encode = make(map[int64]int)

			for i, value := range sortedUniqueValues {
				encode[value] = i
			}

			minV = 0
			gcd = 1
		} else {
			uniqueValues = nil
			// we do blocks if that appears to save 10+% storage
			doBlocks = minMax.spaceInBits > 0 && float64(blockMinMax.spaceInBits)/float64(minMax.spaceInBits) <= 0.9
			if doBlocks {
				numBitsPerValue = 0xFF
				d.meta.WriteUint32(ctx, -2-DV_NUMERIC_BLOCK_SHIFT) // tablesize
			} else {
				numBitsPerValue = packed.UnsignedBitsRequired(uint64((maxV - minV) / gcd))
				if gcd == 1 && minV > 0 && packed.UnsignedBitsRequired(uint64(maxV)) == packed.UnsignedBitsRequired(uint64(maxV-minV)) {
					minV = 0
				}
				d.meta.WriteUint32(ctx, -1) // tablesize
			}
		}
	}

	d.meta.WriteByte(byte(numBitsPerValue))
	d.meta.WriteUint64(ctx, uint64(minV))
	d.meta.WriteUint64(ctx, uint64(gcd))
	startOffset := d.data.GetFilePointer()
	d.meta.WriteUint64(ctx, uint64(startOffset)) // valueOffset
	jumpTableOffset := int64(-1)
	if doBlocks {
		numeric, err := valuesProducer.GetSortedNumeric(ctx, field)
		if err != nil {
			return nil, err
		}
		jumpTableOffset, err = d.writeValuesMultipleBlocks(numeric, gcd)
	} else if numBitsPerValue != 0 {
		numeric, err := valuesProducer.GetSortedNumeric(ctx, field)
		if err != nil {
			return nil, err
		}
		d.writeValuesSingleBlock(numeric, numValues, numBitsPerValue, minV, gcd, encode)
	}
	d.meta.WriteUint64(ctx, uint64(d.data.GetFilePointer()-startOffset)) // valuesLength
	d.meta.WriteUint64(ctx, uint64(jumpTableOffset))
	return []int64{numDocsWithValue, numValues}, nil
}

func (d *DocValuesConsumer) writeValuesMultipleBlocks(values index.SortedNumericDocValues, gcd int64) (int64, error) {
	panic("")
}

func (d *DocValuesConsumer) writeValuesSingleBlock(values index.SortedNumericDocValues, numValues int64, numBitsPerValue int,
	minV int64, gcd int64, encode map[int64]int) error {

	panic("")
}

func GCD(a, b int64) int64 {
	return new(big.Int).GCD(nil, nil, big.NewInt(a), big.NewInt(b)).Int64()
}

type MinMaxTracker struct {
	min, max, numValues, spaceInBits int64
}

func (t MinMaxTracker) update(v int64) {

}

func (t MinMaxTracker) nextBlock() {

}

func (t MinMaxTracker) finish() {

}

func newMinMaxTracker() *MinMaxTracker {
	return &MinMaxTracker{}
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

	return &CompressedBinaryBlockWriter{
		consumer:            d,
		lz4writer:           lz4.NewWriter(d.data),
		tempBinaryOffsets:   tempBinaryOffsets,
		blockAddressesStart: d.data.GetFilePointer(),
	}, nil
}

func (d *DocValuesConsumer) addTermsDict(ctx context.Context, values index.SortedSetDocValues) error {
	panic("implement me")
}

func (d *DocValuesConsumer) compressAndGetTermsDictBlockLength(ctx context.Context, bufferedOutput *store.ByteArrayDataOutput, writer *lz4.Writer) error {
	panic("implement me")
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
