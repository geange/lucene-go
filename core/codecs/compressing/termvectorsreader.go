package compressing

import (
	"bytes"
	"context"
	"iter"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
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
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) Clone(ctx context.Context) index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) GetMergeInstance() index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
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
	fieldNums, fieldFlags, fieldNumOffs, numTerms, fieldLengths                                            []int
	prefixLengths, suffixLengths, termFreqs, positionIndex, positions, startOffsets, lengths, payloadIndex [][]int
	suffixBytes, payloadBytes                                                                              *bytes.Buffer
}

func (f *TVFields) Names() []string {
	//TODO implement me
	panic("implement me")
}

func (f *TVFields) Iterator() iter.Seq[string] {
	panic("implement me")
}

func (f *TVFields) Terms(field string) (index.Terms, error) {
	//TODO implement me
	panic("implement me")
}

func (f *TVFields) Size() int {
	//TODO implement me
	panic("implement me")
}
