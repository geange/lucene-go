package compressing

import (
	"context"
	"iter"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
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
	lastTerm          []byte
	positionsBuf      []int
	startOffsetsBuf   []int
	lengthsBuf        []int
	payloadLengthsBuf []int
	termSuffixes      *store.BufferDataOutput // buffered term suffixes
	payloadBytes      *store.BufferDataOutput // buffered term payloads
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
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) StartField(ctx context.Context, fieldInfo *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) FinishField(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) StartTerm(ctx context.Context, term []byte, freq int) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) FinishTerm(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) AddPosition(ctx context.Context, position, startOffset, endOffset int, payload []byte) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) Finish(ctx context.Context, fieldInfos index.FieldInfos, numDocs int) error {
	//TODO implement me
	panic("implement me")
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
	hasPositions, hasOffsets, hasPayloads bool
	fieldNum, flags, numTerms             int
	freqs, prefixLengths, suffixLengths   []int
	posStart, offStart, payStart          int
	totalPositions                        int
	ord                                   int
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
