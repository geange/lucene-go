package compressing

import (
	"bytes"
	"context"
	"iter"

	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/codecs"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/automaton"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ index.TermVectorsReader = &TermVectorsReader{}

type TermVectorsReader struct {
	fieldInfos        index.FieldInfos
	indexReader       FieldsIndex
	vectorsStream     store.IndexInput
	version           int
	packedIntsVersion int
	compressionMode   CompressionMode
	decompressor      Decompressor
	chunkSize         int
	numDocs           int
	closed            bool
	reader            *packed.BlockPackedReaderIterator
	numChunks         int // number of written blocks
	numDirtyChunks    int // number of incomplete compressed blocks written
	numDirtyDocs      int // cumulative number of docs in incomplete chunks
	maxPointer        int // end of the data section
}

func (t *TermVectorsReader) Close() error {
	t.closed = true
	return nil
}

func (t *TermVectorsReader) Get(ctx context.Context, doc int) (index.Fields, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) CheckIntegrity() error {
	if err := t.indexReader.CheckIntegrity(); err != nil {
		return err
	}
	_, err := codecs.ChecksumEntireFile(context.Background(), t.vectorsStream)
	return err
}

func (t *TermVectorsReader) Clone(ctx context.Context) index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) GetMergeInstance() index.TermVectorsReader {
	return t
}

func (t *TermVectorsReader) GetCompressionMode() CompressionMode {
	return t.compressionMode
}

func (t *TermVectorsReader) GetChunkSize() int {
	return t.chunkSize
}

func (t *TermVectorsReader) GetPackedIntsVersion() int {
	return t.packedIntsVersion
}

func (t *TermVectorsReader) GetVersion() int {
	return t.version
}

func (t *TermVectorsReader) GetIndexReader() FieldsIndex {
	return t.indexReader
}

func (t *TermVectorsReader) GetVectorsStream() store.IndexInput {
	return t.vectorsStream
}

func (t *TermVectorsReader) GetMaxPointer() int {
	return t.maxPointer
}

func (t *TermVectorsReader) GetNumDirtyDocs() int {
	return t.numDirtyDocs
}

func (t *TermVectorsReader) GetNumDirtyChunks() int {
	return t.numDirtyChunks
}

func (t *TermVectorsReader) GetNumChunks() int {
	return t.numChunks
}

func (t *TermVectorsReader) GetNumDocs() int {
	return t.numDocs
}

type Fields interface {
	index.Fields

	// Iterator
	// Returns an iterator that will step through all fields names. This will not return null.
	Iterator() iter.Seq[string]
}

var _ index.Fields = &TVFields{}

type TVFields struct {
	*TermVectorsReader

	fieldNums                 []int
	fieldFlags                []int
	fieldNumOffs              []int
	numTerms                  []int
	fieldLengths              []int
	prefixLengths             [][]int
	suffixLengths             [][]int
	termFreqs                 [][]int
	positionIndex             [][]int
	positions                 [][]int
	startOffsets              [][]int
	lengths                   [][]int
	payloadIndex              [][]int
	suffixBytes, payloadBytes *bytes.Buffer
}

func (t *TermVectorsReader) NewTVFields() *TVFields {
	return nil
}

func (f *TVFields) Names() []string {
	names := make([]string, 0)
	for name := range f.Iterator() {
		names = append(names, name)
	}
	return names
}

func (f *TVFields) Iterator() iter.Seq[string] {
	return func(yield func(string) bool) {
		size := len(f.fieldNumOffs)
		for i := 0; i < size; i++ {
			fieldNum := f.fieldNums[f.fieldNumOffs[i]]
			name := f.fieldInfos.FieldInfoByNumber(fieldNum).Name()
			if !yield(name) {
				return
			}
		}
	}
}

func (f *TVFields) Terms(field string) (index.Terms, error) {
	fieldInfo := f.fieldInfos.FieldInfo(field)
	if fieldInfo == nil {
		return nil, nil
	}
	idx := -1
	for i := 0; i < len(f.fieldNumOffs); i++ {
		if f.fieldNums[f.fieldNumOffs[i]] == fieldInfo.Number() {
			idx = i
			break
		}
	}

	if idx == -1 || f.numTerms[idx] == 0 {
		// no term
		return nil, nil
	}

	fieldOff := 0
	//fieldLen := -1
	for i := 0; i < len(f.fieldNumOffs); i++ {
		if i < idx {
			fieldOff += f.fieldLengths[i]
		} else {
			// TODO: fix it
			//fieldLen = f.fieldLengths[i]
			break
		}
	}

	return NewTVTerms(f.numTerms[idx], f.fieldFlags[idx],
		f.prefixLengths[idx], f.suffixLengths[idx], f.termFreqs[idx],
		f.positionIndex[idx], f.positions[idx], f.startOffsets[idx], f.lengths[idx],
		f.payloadIndex[idx], f.payloadBytes, new(bytes.Buffer)), nil
}

func (f *TVFields) Size() int {
	return len(f.fieldNumOffs)
}

var _ index.Terms = &TVTerms{}

type TVTerms struct {
	numTerms      int
	flags         int
	totalTermFreq int
	prefixLengths []int
	suffixLengths []int
	termFreqs     []int
	positionIndex []int
	positions     []int
	startOffsets  []int
	lengths       []int
	payloadIndex  []int
	termBytes     *bytes.Buffer
	payloadBytes  *bytes.Buffer
}

func NewTVTerms(numTerms int, flags int,
	prefixLengths []int, suffixLengths []int, termFreqs []int, positionIndex []int,
	positions []int, startOffsets []int, lengths []int, payloadIndex []int,
	termBytes *bytes.Buffer, payloadBytes *bytes.Buffer) *TVTerms {
	return &TVTerms{
		numTerms:      numTerms,
		flags:         flags,
		totalTermFreq: lo.Sum(termFreqs),
		prefixLengths: prefixLengths,
		suffixLengths: suffixLengths,
		termFreqs:     termFreqs,
		positionIndex: positionIndex,
		positions:     positions,
		startOffsets:  startOffsets,
		lengths:       lengths,
		payloadIndex:  payloadIndex,
		termBytes:     termBytes,
		payloadBytes:  payloadBytes,
	}

}

func (t *TVTerms) Iterator() (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) HasFreqs() bool {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) HasOffsets() bool {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) HasPositions() bool {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) HasPayloads() bool {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) GetMin() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTerms) GetMax() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.TermsEnum = &TVTermsEnum{}

type TVTermsEnum struct {
	*coreIndex.BaseTermsEnum

	numTerms      int
	startPos      int
	ord           int
	prefixLengths []int
	suffixLengths []int
	termFreqs     []int
	positionIndex []int
	positions     []int
	startOffsets  []int
	lengths       []int
	payloadIndex  []int
	in            *store.BufferInput
	payloads      *bytes.Buffer
	term          *bytes.Buffer
}

func NewTVTermsEnum() *TVTermsEnum {
	return &TVTermsEnum{
		// TODO：reuse buffer
		term: new(bytes.Buffer),
	}
}

func (t *TVTermsEnum) Next(ctx context.Context) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.PostingsEnum = &TVPostingsEnum{}

type TVPostingsEnum struct {
	doc               int
	termFreq          int
	positionIndex     int
	positions         []int
	startOffsets      []int
	lengths           []int
	payload           *bytes.Buffer
	payloadIndex      []int
	basePayloadOffset int
	i                 int
}

func NewTVPostingsEnum() *TVPostingsEnum {
	return &TVPostingsEnum{
		doc:     -1,
		payload: new(bytes.Buffer),
	}
}

func (t *TVPostingsEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TVPostingsEnum) GetPayload() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
