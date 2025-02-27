package compressing

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"math"
	"slices"
	"sort"

	"github.com/geange/gods-generic/sets/treeset"
	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/array"
	"github.com/geange/lucene-go/core/util/packed"
)

const (
	VECTORS_EXTENSION        = "tvd"
	VECTORS_INDEX_EXTENSION  = "tvx"
	VECTORS_META_EXTENSION   = "tvm"
	VECTORS_INDEX_CODEC_NAME = "Lucene85TermVectorsIndex"

	VECTORS_VERSION_START         = 1
	VECTORS_VERSION_OFFHEAP_INDEX = 2 // Version where all metadata were moved to the meta file.
	VECTORS_VERSION_META          = 3
	VECTORS_VERSION_NUMCHUNKS     = 4 // Version where numChunks is explicitly recorded in meta file
	VECTORS_VERSION_CURRENT       = VECTORS_VERSION_NUMCHUNKS
	VECTORS_META_VERSION_START    = 0

	VECTORS_PACKED_BLOCK_SIZE = 64

	VECTORS_POSITIONS = 0x01
	VECTORS_OFFSETS   = 0x02
	VECTORS_PAYLOADS  = 0x04
)

var (
	VECTORS_FLAGS_BITS, _ = packed.BitsRequired(VECTORS_POSITIONS | VECTORS_OFFSETS | VECTORS_PAYLOADS)
)

var _ index.TermVectorsWriter = &TermVectorsWriter{}

type TermVectorsWriter struct {
	segment         string
	indexWriter     *FieldsIndexWriter
	metaStream      store.IndexOutput
	vectorsStream   store.IndexOutput
	compressionMode CompressionMode
	compressor      Compressor
	chunkSize       int
	numChunks       int // number of chunks
	numDirtyChunks  int // number of incomplete compressed blocks written
	numDirtyDocs    int // cumulative number of docs in incomplete chunks

	numDocs           int              // total number of docs seen
	pendingDocs       *Deque[*DocData] // pending docs
	curDoc            *DocData         // current document
	curField          *FieldData       // current field
	lastTerm          *bytes.Buffer
	positionsBuf      []int
	startOffsetsBuf   []int
	lengthsBuf        []int
	payloadLengthsBuf []int
	termSuffixes      *bytes.Buffer // buffered term suffixes
	payloadBytes      *bytes.Buffer // buffered term payloads
	writer            *packed.BlockPackedWriter
	maxDocsPerChunk   int // hard limit on number of docs per chunk
}

func (t *TermVectorsWriter) Close() error {
	if err := t.metaStream.Close(); err != nil {
		return err
	}
	if err := t.vectorsStream.Close(); err != nil {
		return err
	}
	if err := t.indexWriter.Close(); err != nil {
		return err
	}
	t.metaStream = nil
	t.vectorsStream = nil
	t.indexWriter = nil
	return nil
}

func (t *TermVectorsWriter) StartDocument(ctx context.Context, numVectorFields int) error {
	t.curDoc = t.addDocData(numVectorFields)
	return nil
}

func (t *TermVectorsWriter) FinishDocument(ctx context.Context) error {
	// append the payload bytes of the doc after its terms
	t.termSuffixes.Write(t.payloadBytes.Bytes())
	t.payloadBytes.Reset()
	t.numDocs++
	if t.triggerFlush() {
		if err := t.flush(ctx); err != nil {
			return err
		}
	}
	t.curDoc = nil
	return nil
}

func (t *TermVectorsWriter) StartField(ctx context.Context, info *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error {
	t.curField = t.curDoc.addField(info.Number(), numTerms, positions, offsets, payloads)
	t.lastTerm.Reset()
	return nil
}

func (t *TermVectorsWriter) FinishField(ctx context.Context) error {
	t.curField = nil
	return nil
}

func (t *TermVectorsWriter) StartTerm(ctx context.Context, term []byte, freq int) error {
	prefix := 0
	if t.lastTerm.Len() > 0 {
		diff, err := util.BytesDifference(t.lastTerm.Bytes(), term)
		if err != nil {
			return err
		}
		prefix = diff
	}

	t.curField.addTerm(freq, prefix, len(term)-prefix)
	t.termSuffixes.Write(term[prefix:])
	t.lastTerm.Reset()
	t.lastTerm.Write(term)
	return nil
}

func (t *TermVectorsWriter) FinishTerm(ctx context.Context) error {
	return nil
}

func (t *TermVectorsWriter) AddPosition(ctx context.Context, position, startOffset, endOffset int, payload []byte) error {

	payloadLength := 0
	if len(payload) > 0 {
		payloadLength = len(payload)
	}
	t.curField.addPosition(position, startOffset, endOffset-startOffset, payloadLength)
	if t.curField.hasPayloads && len(payload) > 0 {
		t.payloadBytes.Write(payload)
	}
	return nil
}

func (t *TermVectorsWriter) Finish(ctx context.Context, fieldInfos index.FieldInfos, numDocs int) error {
	if !t.pendingDocs.Empty() {
		t.numDirtyChunks++ // incomplete: we had to force this flush
		t.numDirtyDocs += t.pendingDocs.Size()
		if err := t.flush(ctx); err != nil {
			return err
		}
	}

	if numDocs != t.numDocs {
		return fmt.Errorf("wrote %d docs, finish called with numDocs=%d", t.numDocs, numDocs)
	}

	if err := t.indexWriter.finish(ctx, numDocs, t.vectorsStream.GetFilePointer(), t.metaStream); err != nil {
		return err
	}
	if err := t.metaStream.WriteUvarint(ctx, uint64(t.numChunks)); err != nil {
		return err
	}
	if err := t.metaStream.WriteUvarint(ctx, uint64(t.numDirtyChunks)); err != nil {
		return err
	}
	if err := t.metaStream.WriteUvarint(ctx, uint64(t.numDirtyDocs)); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, t.metaStream); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, t.vectorsStream); err != nil {
		return err
	}
	return nil
}

func (t *TermVectorsWriter) AddProx(numProx int, positions, offsets store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) triggerFlush() bool {
	return t.termSuffixes.Len() >= t.chunkSize || t.pendingDocs.Size() >= t.maxDocsPerChunk
}

func (t *TermVectorsWriter) flush(ctx context.Context) error {
	t.numChunks++
	chunkDocs := t.pendingDocs.Size()

	// write the index file
	if err := t.indexWriter.writeIndex(ctx, chunkDocs, t.vectorsStream.GetFilePointer()); err != nil {
		return err
	}

	docBase := t.numDocs - chunkDocs
	if err := t.vectorsStream.WriteUvarint(ctx, uint64(docBase)); err != nil {
		return err
	}
	if err := t.vectorsStream.WriteUvarint(ctx, uint64(chunkDocs)); err != nil {
		return err
	}

	// total number of fields of the chunk
	totalFields, err := t.flushNumFieldsWithChunkDocs(ctx, chunkDocs)
	if err != nil {
		return err
	}

	if totalFields > 0 {
		// unique field numbers (sorted)
		fieldNums, err := t.flushFieldNums(ctx)
		if err != nil {
			return err
		}
		// offsets in the array of unique field numbers
		if err := t.flushFields(totalFields, fieldNums); err != nil {
			return err
		}
		// flags (does the field have positions, offsets, payloads?)
		if err := t.flushFlags(ctx, totalFields, fieldNums); err != nil {
			return err
		}
		// number of terms of each field
		if err := t.flushNumTerms(ctx, totalFields); err != nil {
			return err
		}
		// prefix and suffix lengths for each field
		if err := t.flushTermLengths(ctx); err != nil {
			return err
		}
		// term freqs - 1 (because termFreq is always >=1) for each term
		if err := t.flushTermFreqs(ctx); err != nil {
			return err
		}
		// positions for all terms, when enabled
		if err := t.flushPositions(ctx); err != nil {
			return err
		}
		// offsets for all terms, when enabled
		if err := t.flushOffsets(ctx, fieldNums); err != nil {
			return err
		}
		// payload lengths for all terms, when enabled
		if err := t.flushPayloadLengths(ctx); err != nil {
			return err
		}

		// compress terms and payloads and write them to the output
		if err := t.compressor.Compress(ctx, t.termSuffixes.Bytes(), t.vectorsStream); err != nil {
			return err
		}
	}

	// reset
	t.pendingDocs.Clear()
	t.curDoc = nil
	t.curField = nil
	t.termSuffixes.Reset()
	return nil
}

func (t *TermVectorsWriter) flushNumFieldsWithChunkDocs(ctx context.Context, chunkDocs int) (int, error) {
	if chunkDocs == 1 {
		numFields := t.pendingDocs.First().numFields
		if err := t.vectorsStream.WriteUvarint(ctx, uint64(numFields)); err != nil {
			return 0, err
		}
		return numFields, nil
	} else {
		t.writer.Reset(t.vectorsStream)
		totalFields := 0
		for dd := range t.pendingDocs.Iterator() {
			if err := t.writer.Add(ctx, uint64(dd.numFields)); err != nil {
				return 0, err
			}
			totalFields += dd.numFields
		}
		if err := t.writer.Finish(ctx); err != nil {
			return 0, err
		}
		return totalFields, nil
	}
}

// Returns a sorted array containing unique field numbers
func (t *TermVectorsWriter) flushFieldNums(ctx context.Context) ([]int, error) {
	fieldNums := treeset.New[int]()
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			fieldNums.Add(fd.fieldNum)
		}
	}

	numDistinctFields := fieldNums.Size()
	bitsRequired, err := packed.BitsRequired(int64(fieldNums.Size()))
	if err != nil {
		return nil, err
	}
	token := (min(numDistinctFields-1, 0x07) << 5) | bitsRequired
	if err := t.vectorsStream.WriteByte(byte(token)); err != nil {
		return nil, err
	}
	if numDistinctFields-1 >= 0x07 {
		if err := t.vectorsStream.WriteUvarint(ctx, uint64(numDistinctFields-1-0x07)); err != nil {
			return nil, err
		}
	}

	writer := packed.GetWriterNoHeader(t.vectorsStream, packed.FormatPacked, fieldNums.Size(), bitsRequired, 1)
	fValues := fieldNums.Values()
	for _, fieldNum := range fValues {
		if err := writer.Add(uint64(fieldNum)); err != nil {
			return nil, err
		}
	}
	if err := writer.Finish(); err != nil {
		return nil, err
	}

	fns := make([]int, len(fValues))
	for i, fieldNum := range fValues {
		fns[i] = fieldNum
	}
	return fns, nil
}

func (t *TermVectorsWriter) flushFields(totalFields int, fieldNums []int) error {
	bitsRequired, err := packed.BitsRequired(int64(len(fieldNums) - 1))
	if err != nil {
		return err
	}
	writer := packed.GetWriterNoHeader(t.vectorsStream, packed.FormatPacked, totalFields, bitsRequired, 1)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			fieldNumIndex := sort.SearchInts(fieldNums, fd.fieldNum)
			if err := writer.Add(uint64(fieldNumIndex)); err != nil {
				return err
			}
		}
	}
	return writer.Finish()
}

func (t *TermVectorsWriter) flushFlags(ctx context.Context, totalFields int, fieldNums []int) error {
	// check if fields always have the same flags
	nonChangingFlags := true
	fieldFlags := make([]int, len(fieldNums))
	for i := range fieldFlags {
		fieldFlags[i] = -1
	}

OUTER:
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			fieldNumOff, ok := slices.BinarySearch(fieldNums, fd.fieldNum)
			if !ok {
				continue
			}
			if fieldFlags[fieldNumOff] == -1 {
				fieldFlags[fieldNumOff] = fd.flags
			} else if fieldFlags[fieldNumOff] != fd.flags {
				nonChangingFlags = false
				break OUTER
			}
		}
	}

	if nonChangingFlags {
		// write one flag per field num
		if err := t.vectorsStream.WriteUvarint(ctx, 0); err != nil {
			return err
		}
		writer := packed.GetWriterNoHeader(t.vectorsStream, packed.FormatPacked, len(fieldFlags), VECTORS_FLAGS_BITS, 1)
		for _, flags := range fieldFlags {
			if err := writer.Add(uint64(flags)); err != nil {
				return err
			}
		}
		return writer.Finish()
	}

	// write one flag for every field instance
	if err := t.vectorsStream.WriteUvarint(ctx, 1); err != nil {
		return err
	}
	writer := packed.GetWriterNoHeader(t.vectorsStream, packed.FormatPacked, totalFields, VECTORS_FLAGS_BITS, 1)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			if err := writer.Add(uint64(fd.flags)); err != nil {
				return err
			}
		}
	}
	return writer.Finish()
}

func (t *TermVectorsWriter) flushNumTerms(ctx context.Context, totalFields int) error {
	maxNumTerms := 0
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			maxNumTerms |= fd.numTerms
		}
	}
	bitsRequired, err := packed.BitsRequired(int64(maxNumTerms))
	if err != nil {
		return err
	}
	if err := t.vectorsStream.WriteUvarint(ctx, uint64(bitsRequired)); err != nil {
		return err
	}
	writer := packed.GetWriterNoHeader(t.vectorsStream, packed.FormatPacked, totalFields, bitsRequired, 1)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			if err := writer.Add(uint64(fd.numTerms)); err != nil {
				return err
			}
		}
	}
	return writer.Finish()
}

func (t *TermVectorsWriter) flushTermLengths(ctx context.Context) error {
	t.writer.Reset(t.vectorsStream)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			for i := 0; i < fd.numTerms; i++ {
				if err := t.writer.Add(ctx, uint64(fd.prefixLengths[i])); err != nil {
					return err
				}
			}
		}
	}
	if err := t.writer.Finish(ctx); err != nil {
		return err
	}
	t.writer.Reset(t.vectorsStream)

	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			for i := 0; i < fd.numTerms; i++ {
				if err := t.writer.Add(ctx, uint64(fd.suffixLengths[i])); err != nil {
					return err
				}
			}
		}
	}
	return t.writer.Finish(ctx)
}

func (t *TermVectorsWriter) flushTermFreqs(ctx context.Context) error {
	t.writer.Reset(t.vectorsStream)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			for i := 0; i < fd.numTerms; i++ {
				if err := t.writer.Add(ctx, uint64(fd.freqs[i]-1)); err != nil {
					return err
				}
			}
		}
	}
	return t.writer.Finish(ctx)
}

func (t *TermVectorsWriter) flushPositions(ctx context.Context) error {
	t.writer.Reset(t.vectorsStream)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			pos := 0
			for i := 0; i < fd.numTerms; i++ {
				previousPosition := 0
				for j := 0; j < fd.freqs[i]; j++ {
					position := t.positionsBuf[fd.posStart+pos]
					pos++
					if err := t.writer.Add(ctx, uint64(position-previousPosition)); err != nil {
						return err
					}
					previousPosition = position
				}
			}
		}
	}
	return t.writer.Finish(ctx)
}

func (t *TermVectorsWriter) flushOffsets(ctx context.Context, fieldNums []int) error {
	hasOffsets := false
	sumPos := make([]int, len(fieldNums))
	sumOffsets := make([]int, len(fieldNums))
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			hasOffsets = hasOffsets || fd.hasOffsets
			if fd.hasOffsets && fd.hasPositions {
				fieldNumOff, found := slices.BinarySearch(fieldNums, fd.fieldNum)
				if found {
					pos := 0
					for i := 0; i < fd.numTerms; i++ {
						sumPos[fieldNumOff] += t.positionsBuf[fd.posStart+fd.freqs[i]-1+pos]
						sumOffsets[fieldNumOff] += t.startOffsetsBuf[fd.offStart+fd.freqs[i]-1+pos]
						pos += fd.freqs[i]
					}
				}
			}
		}
	}

	if !hasOffsets {
		return nil
	}

	charsPerTerm := make([]float32, len(fieldNums))
	for i := 0; i < len(fieldNums); i++ {
		charsPerTerm[i] = 0
		if !(sumPos[i] <= 0 || sumOffsets[i] <= 0) {
			charsPerTerm[i] = float32(sumOffsets[i]) / float32(sumPos[i])
		}

		// start offsets
		num := math.Float32bits(charsPerTerm[i])
		if err := t.vectorsStream.WriteUint32(ctx, num); err != nil {
			return err
		}
	}

	//// start offsets
	//for i := 0; i < len(fieldNums); i++ {
	//	num := math.Float32bits(charsPerTerm[i])
	//	if err := t.vectorsStream.WriteUint32(ctx, num); err != nil {
	//		return err
	//	}
	//}

	t.writer.Reset(t.vectorsStream)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			if (fd.flags & VECTORS_OFFSETS) != 0 {
				fieldNumOff, found := slices.BinarySearch(fieldNums, fd.fieldNum)
				if !found {
					continue
				}
				cpt := charsPerTerm[fieldNumOff]
				pos := 0

				for i := 0; i < fd.numTerms; i++ {
					previousPos := 0
					previousOff := 0

					for j := 0; j < fd.freqs[i]; j++ {
						position := 0
						if !fd.hasOffsets {
							position = t.positionsBuf[fd.posStart+pos]
						}
						startOffset := t.startOffsetsBuf[fd.offStart+pos]

						num := startOffset - previousOff - int(cpt*float32(position-previousPos))
						if err := t.writer.Add(ctx, uint64(num)); err != nil {
							return err
						}
						previousPos = position
						previousOff = startOffset
						pos++
					}
				}
			}
		}
	}
	if err := t.writer.Finish(ctx); err != nil {
		return err
	}

	// lengths
	t.writer.Reset(t.vectorsStream)
	for dd := range t.pendingDocs.Iterator() {
		for fd := range dd.fields.Iterator() {
			if (fd.flags & VECTORS_OFFSETS) != 0 {
				pos := 0
				for i := 0; i < fd.numTerms; i++ {
					for j := 0; j < fd.freqs[i]; j++ {
						n := t.lengthsBuf[fd.offStart+pos] - fd.prefixLengths[i] - fd.suffixLengths[i]
						if err := t.writer.Add(ctx, uint64(n)); err != nil {
							return err
						}
						pos++
					}
				}
			}
		}
	}
	return t.writer.Finish(ctx)
}

func (t *TermVectorsWriter) flushPayloadLengths(ctx context.Context) error {
	t.writer.Reset(t.vectorsStream)
	for docData := range t.pendingDocs.Iterator() {
		for fieldData := range docData.fields.Iterator() {
			if fieldData.hasPayloads {
				for i := 0; i < fieldData.totalPositions; i++ {
					if err := t.writer.Add(ctx, uint64(t.payloadLengthsBuf[fieldData.payStart+i])); err != nil {
						return err
					}
				}
			}
		}
	}
	return t.writer.Finish(ctx)
}

// DocData
// a pending doc
type DocData struct {
	numFields int
	fields    *Deque[*FieldData]
	posStart  int
	offStart  int
	payStart  int
}

func newDocData(numFields int, posStart int, offStart int, payStart int) *DocData {
	return &DocData{
		numFields: numFields,
		fields:    NewDeque[*FieldData](),
		posStart:  posStart,
		offStart:  offStart,
		payStart:  payStart,
	}
}

func (d *DocData) addField(fieldNum, numTerms int, positions, offsets, payloads bool) *FieldData {
	var field *FieldData
	if d.fields.Size() == 0 {
		field = newFieldData(fieldNum, numTerms, positions, offsets, payloads, d.posStart, d.offStart, d.payStart)
	} else {
		last := d.fields.Last()
		posStart := last.posStart
		if last.hasPositions {
			posStart += last.totalPositions
		}

		offStart := last.offStart
		if last.hasOffsets {
			offStart += last.totalPositions
		}

		payStart := last.payStart
		if last.hasPayloads {
			payStart += last.totalPositions
		}

		field = newFieldData(fieldNum, numTerms, positions, offsets, payloads, posStart, offStart, payStart)
	}
	d.fields.Add(field)
	return field
}

func (t *TermVectorsWriter) addDocData(numVectorFields int) *DocData {
	var last *FieldData
	for doc := range t.pendingDocs.DescIterator() {
		if !doc.fields.Empty() {
			last = doc.fields.Last()
			break
		}
	}

	var doc *DocData
	if last == nil {
		doc = newDocData(numVectorFields, 0, 0, 0)
	} else {
		posStart := last.posStart
		if last.hasPositions {
			posStart += last.totalPositions
		}

		offStart := last.offStart
		if last.hasOffsets {
			offStart += last.totalPositions
		}

		payStart := last.payStart
		if last.hasPayloads {
			payStart += last.totalPositions
		}

		doc = newDocData(numVectorFields, posStart, offStart, payStart)
	}
	t.pendingDocs.Add(doc)
	return doc
}

// FieldData a pending field
type FieldData struct {
	writer                                *TermVectorsWriter
	hasPositions, hasOffsets, hasPayloads bool
	fieldNum, flags, numTerms             int
	freqs, prefixLengths, suffixLengths   []int
	posStart, offStart, payStart          int
	totalPositions                        int
	ord                                   int
}

func (f *FieldData) addTerm(freq, prefixLength, suffixLength int) {
	f.freqs[f.ord] = freq
	f.prefixLengths[f.ord] = prefixLength
	f.suffixLengths[f.ord] = suffixLength
	f.ord++
}

func (f *FieldData) addPosition(position, startOffset, length, payloadLength int) {
	if f.hasPositions {
		if f.posStart+f.totalPositions == len(f.writer.positionsBuf) {
			f.writer.positionsBuf = slices.Grow(f.writer.positionsBuf, 1)
		}
		f.writer.positionsBuf[f.posStart+f.totalPositions] = position
	}
	if f.hasOffsets {
		if f.offStart+f.totalPositions == len(f.writer.startOffsetsBuf) {
			newLength := array.Oversize(f.offStart+f.totalPositions, 4)
			f.writer.startOffsetsBuf = array.GrowExact(f.writer.startOffsetsBuf, newLength)
			f.writer.lengthsBuf = array.GrowExact(f.writer.lengthsBuf, newLength)
		}
		f.writer.startOffsetsBuf[f.offStart+f.totalPositions] = startOffset
		f.writer.lengthsBuf[f.offStart+f.totalPositions] = length
	}
	if f.hasPayloads {
		if f.payStart+f.totalPositions == len(f.writer.payloadLengthsBuf) {
			f.writer.payloadLengthsBuf = slices.Grow(f.writer.payloadLengthsBuf, 1)
		}
		f.writer.payloadLengthsBuf[f.payStart+f.totalPositions] = payloadLength
	}
	f.totalPositions++
}

func newFieldData(fieldNum, numTerms int, positions, offsets, payloads bool,
	posStart, offStart, payStart int) *FieldData {

	flags := 0
	if positions {
		flags = flags | VECTORS_POSITIONS
	}
	if offsets {
		flags = flags | VECTORS_OFFSETS
	}
	if payloads {
		flags = flags | VECTORS_PAYLOADS
	}

	return &FieldData{
		hasPositions:   positions,
		hasOffsets:     offsets,
		hasPayloads:    payloads,
		fieldNum:       fieldNum,
		flags:          flags,
		numTerms:       numTerms,
		freqs:          make([]int, numTerms),
		prefixLengths:  make([]int, numTerms),
		suffixLengths:  make([]int, numTerms),
		posStart:       posStart,
		offStart:       offStart,
		payStart:       payStart,
		totalPositions: 0,
		ord:            0,
	}
}

type Deque[T any] struct {
	values []T
}

func NewDeque[T any]() *Deque[T] {
	return &Deque[T]{values: make([]T, 0)}
}

func (d *Deque[T]) Empty() bool {
	return len(d.values) == 0
}

func (d *Deque[T]) Size() int {
	return len(d.values)
}

func (d *Deque[T]) Add(item T) {
	d.values = append(d.values, item)
}

func (d *Deque[T]) Clear() {
	d.values = d.values[:0]
}

func (d *Deque[T]) First() T {
	return d.values[len(d.values)-1]
}

func (d *Deque[T]) Last() T {
	return d.values[0]
}

func (d *Deque[T]) DescIterator() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < len(d.values); i++ {
			if !yield(d.values[i]) {
				return
			}
		}
	}
}

func (d *Deque[T]) Iterator() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(d.values) - 1; i >= 0; i-- {
			if !yield(d.values[i]) {
				return
			}
		}
	}
}
