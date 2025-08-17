package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ index.DocValuesProducer = &DocValuesProducer{}

type DocValuesProducer struct {
	numerics       map[string]*NumericEntry
	binaries       map[string]*BinaryEntry
	sorted         map[string]*SortedEntry
	sortedSets     map[string]*SortedSetEntry
	sortedNumerics map[string]*SortedNumericEntry
	ramBytesUsed   int
	data           store.IndexInput
	maxDoc         int
	version        int
}

func NewDocValuesProducer(ctx context.Context, state *index.SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (*DocValuesProducer, error) {
	panic("implement me")
}

func (d *DocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetNumeric(ctx context.Context, field *document.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetBinary(ctx context.Context, field *document.FieldInfo) (index.BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (index.SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetSortedNumeric(ctx context.Context, field *document.FieldInfo) (index.SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetSortedSet(ctx context.Context, field *document.FieldInfo) (index.SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetMergeInstance() index.DocValuesProducer {
	//TODO implement me
	panic("implement me")
}

type NumericEntry struct {
	table                []int64
	blockShift           int
	bitsPerValue         byte
	docsWithFieldOffset  int64
	docsWithFieldLength  int64
	jumpTableEntryCount  int
	denseRankPower       byte
	numValues            int
	minValue             int64
	gcd                  int64
	valuesOffset         int64
	valuesLength         int
	valueJumpTableOffset int64 // -1 if no jump-table
}

type BinaryEntry struct {
	compressed               bool
	dataOffset               int64
	dataLength               int
	docsWithFieldOffset      int64
	docsWithFieldLength      int64
	jumpTableEntryCount      int
	denseRankPower           byte
	numDocsWithField         int
	minLength                int
	maxLength                int
	addressesOffset          int64
	addressesLength          int
	addressesMeta            *packed.Meta
	numCompressedChunks      int
	docsPerChunkShift        int
	maxUncompressedChunkSize int
}

type TermsDictEntry struct {
	termsDictSize             int
	termsDictBlockShift       int
	termsAddressesMeta        *packed.Meta
	maxTermLength             int
	termsDataOffset           int64
	termsDataLength           int
	termsAddressesOffset      int64
	termsAddressesLength      int
	termsDictIndexShift       int
	termsIndexAddressesMeta   *packed.Meta
	termsIndexOffset          int64
	termsIndexLength          int
	termsIndexAddressesOffset int64
	termsIndexAddressesLength int
	compressed                bool
	maxBlockLength            int
}

type SortedEntry struct {
	TermsDictEntry

	docsWithFieldOffset int64
	docsWithFieldLength int
	jumpTableEntryCount int
	denseRankPower      byte
	numDocsWithField    int
	bitsPerValue        byte
	ordsOffset          int64
	ordsLength          int
}

type SortedSetEntry struct {
	TermsDictEntry

	singleValueEntry    *SortedEntry
	docsWithFieldOffset int64
	docsWithFieldLength int
	jumpTableEntryCount int
	denseRankPower      byte
	numDocsWithField    int
	bitsPerValue        byte
	ordsOffset          int64
	ordsLength          int
	addressesMeta       *packed.Meta
	addressesOffset     int64
	addressesLength     int
}

type SortedNumericEntry struct {
	NumericEntry

	numDocsWithField int
	addressesMeta    *packed.Meta
	addressesOffset  int64
	addressesLength  int
}

var _ index.NumericDocValues = &DenseNumericDocValues{}

type DenseNumericDocValues struct {
}

func (d *DenseNumericDocValues) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (d *DenseNumericDocValues) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DenseNumericDocValues) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DenseNumericDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DenseNumericDocValues) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (d *DenseNumericDocValues) AdvanceExact(target int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DenseNumericDocValues) LongValue() (int64, error) {
	//TODO implement me
	panic("implement me")
}

//var _ index.BinaryDocValues = &DenseBinaryDocValues{}

type DenseBinaryDocValues struct {
	maxDoc int
	doc    int
}

var _ index.NumericDocValues = &SparseNumericDocValues{}

type SparseNumericDocValues struct {
}

func (s *SparseNumericDocValues) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (s *SparseNumericDocValues) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseNumericDocValues) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseNumericDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseNumericDocValues) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SparseNumericDocValues) AdvanceExact(target int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseNumericDocValues) LongValue() (int64, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.BinaryDocValues = &SparseBinaryDocValues{}

type SparseBinaryDocValues struct {
	disi *IndexedDISI
}

func (s *SparseBinaryDocValues) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (s *SparseBinaryDocValues) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseBinaryDocValues) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseBinaryDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseBinaryDocValues) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SparseBinaryDocValues) AdvanceExact(target int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SparseBinaryDocValues) BinaryValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
