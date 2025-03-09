package compressing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"slices"

	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/codecs"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
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

func _newTermVectorsReader(reader *TermVectorsReader) (*TermVectorsReader, error) {
	indexReader, err := reader.indexReader.Clone()
	if err != nil {
		return nil, err
	}

	vectorsStream := reader.vectorsStream.Clone().(store.IndexInput)

	iterator := packed.NewBlockPackedReaderIterator(vectorsStream,
		reader.packedIntsVersion, VECTORS_PACKED_BLOCK_SIZE, 0)

	return &TermVectorsReader{
		fieldInfos:        reader.fieldInfos,
		vectorsStream:     reader.vectorsStream.Clone().(store.IndexInput),
		indexReader:       indexReader,
		packedIntsVersion: reader.packedIntsVersion,
		compressionMode:   reader.compressionMode,
		decompressor:      reader.decompressor.Clone(),
		chunkSize:         reader.chunkSize,
		numDocs:           reader.numDocs,
		reader:            iterator,
		version:           reader.version,
		numChunks:         reader.numChunks,
		numDirtyChunks:    reader.numDirtyChunks,
		numDirtyDocs:      reader.numDirtyDocs,
		maxPointer:        reader.maxPointer,
		closed:            false,
	}, nil
}

func (t *TermVectorsReader) Close() error {
	t.closed = true
	return nil
}

func (t *TermVectorsReader) Get(ctx context.Context, doc int) (index.Fields, error) {
	// seek to the right place
	startPointer, err := t.indexReader.GetStartPointer(doc)
	if err != nil {
		return nil, err
	}
	if _, err := t.vectorsStream.Seek(startPointer, io.SeekStart); err != nil {
		return nil, err
	}

	// decode
	// - docBase: first doc ID of the chunk
	// - chunkDocs: number of docs of the chunk
	docBase, err := t.vectorsStream.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	chunkDocs, err := t.vectorsStream.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	if doc < int(docBase) || doc >= int(docBase+chunkDocs) || int(docBase+chunkDocs) > t.numDocs {
		return nil, fmt.Errorf("docBase=%d,chunkDocs=%d,doc=%d", docBase, chunkDocs, doc)
	}

	skip := 0
	numFields := 0
	totalFields := 0

	if chunkDocs == 1 {
		skip = 0
		num, err := t.vectorsStream.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		numFields = int(num)
		totalFields = int(num)
	} else {
		t.reader.Reset(t.vectorsStream, int(chunkDocs))
		sum := 0
		for i := int(docBase); i < doc; i++ {
			n, err := t.reader.Next(ctx)
			if err != nil {
				return nil, err
			}
			sum += int(n)
		}

		skip = sum

		nFields, err := t.reader.Next(ctx)
		if err != nil {
			return nil, err
		}
		numFields = int(nFields)
		sum += numFields
		for i := doc + 1; i < int(docBase+chunkDocs); i++ {
			n, err := t.reader.Next(ctx)
			if err != nil {
				return nil, err
			}
			sum += int(n)
		}
		totalFields = sum
	}

	if numFields == 0 {
		return nil, nil
	}
	// read field numbers that have term vectors
	var fieldNums []int
	{
		token, err := t.vectorsStream.ReadByte()
		if err != nil {
			return nil, err
		}
		bitsPerFieldNum := int(token & 0x1F)
		totalDistinctFields := int(token >> 5)
		if totalDistinctFields == 0x7 {
			num, err := t.vectorsStream.ReadUvarint(ctx)
			if err != nil {
				return nil, err
			}
			totalDistinctFields += int(num)
		}
		totalDistinctFields++
		seq, err := packed.NewPackedReaderIterator(t.vectorsStream, packed.FormatPacked,
			totalDistinctFields, bitsPerFieldNum, 1).Iterator()
		if err != nil {
			return nil, err
		}
		next, stop := iter.Pull(seq)
		defer stop()

		fieldNums = make([]int, 0, totalDistinctFields)
		for i := 0; i < totalDistinctFields; i++ {
			n, ok := next()
			if ok {
				fieldNums = append(fieldNums, int(n))
			}
		}
	}

	// read field numbers and flags
	fieldNumOffs := make([]int, numFields)
	var flags packed.Reader
	{
		bitsPerOff, err := packed.BitsRequired(int64(len(fieldNums) - 1))
		if err != nil {
			return nil, err
		}
		allFieldNumOffs, err := packed.GetReaderNoHeader(ctx, t.vectorsStream, packed.FormatPacked,
			t.packedIntsVersion, totalFields, bitsPerOff)
		if err != nil {
			return nil, err
		}
		n, err := t.vectorsStream.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		switch n {
		case 0:
			fieldFlags, err := packed.GetReaderNoHeader(ctx, t.vectorsStream, packed.FormatPacked,
				t.packedIntsVersion, len(fieldNums), VECTORS_FLAGS_BITS)
			if err != nil {
				return nil, err
			}
			f := packed.GetMutable(totalFields, VECTORS_FLAGS_BITS, packed.COMPACT)
			for i := 0; i < totalFields; i++ {
				fieldNumOff, err := allFieldNumOffs.Get(i)
				if err != nil {
					return nil, err
				}
				fgs, err := fieldFlags.Get(int(fieldNumOff))
				if err != nil {
					return nil, err
				}
				f.Set(i, fgs)
			}
			flags = f
			break
		case 1:
			noHeader, err := packed.GetReaderNoHeader(ctx, t.vectorsStream, packed.FormatPacked,
				t.packedIntsVersion, totalFields, VECTORS_FLAGS_BITS)
			if err != nil {
				return nil, err
			}
			flags = noHeader
		default:
			return nil, errors.New("assertion error")
		}

		for i := 0; i < numFields; i++ {
			fieldNumOff, err := allFieldNumOffs.Get(skip + i)
			if err != nil {
				return nil, err
			}
			fieldNumOffs[i] = int(fieldNumOff)
		}
	}

	// number of terms per field for all fields
	var numTerms packed.Reader
	totalTerms := 0
	{
		bitsRequired, err := t.vectorsStream.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		noHeader, err := packed.GetReaderNoHeader(ctx, t.vectorsStream, packed.FormatPacked,
			t.packedIntsVersion, totalFields, int(bitsRequired))
		if err != nil {
			return nil, err
		}
		numTerms = noHeader
		sum := 0
		for i := 0; i < totalFields; i++ {
			n, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			sum += int(n)
		}
		totalTerms = sum
	}

	// term lengths
	docOff := 0
	docLen := 0
	totalLen := 0
	fieldLengths := make([]int, numFields)
	prefixLengths := make([][]int, numFields)
	suffixLengths := make([][]int, numFields)
	{
		t.reader.Reset(t.vectorsStream, totalTerms)
		// skip
		toSkip := 0
		for i := 0; i < skip; i++ {
			n, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			toSkip += int(n)
		}
		err := t.reader.Skip(ctx, toSkip)
		if err != nil {
			return nil, err
		}
		// read prefix lengths
		for i := 0; i < numFields; i++ {
			termCount, err := numTerms.Get(skip + i)
			if err != nil {
				return nil, err
			}
			fieldPrefixLengths := make([]int, termCount)
			prefixLengths[i] = fieldPrefixLengths
			for j := 0; j < int(termCount); {
				next, err := t.reader.NextSlices(ctx, int(termCount)-j)
				if err != nil {
					return nil, err
				}
				for k := 0; k < len(next); k++ {
					fieldPrefixLengths[j] = int(next[k])
					j++
				}
			}
		}
		err = t.reader.Skip(ctx, totalTerms-t.reader.Ord())
		if err != nil {
			return nil, err
		}

		t.reader.Reset(t.vectorsStream, totalTerms)
		// skip
		toSkip = 0
		for i := 0; i < skip; i++ {
			n, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			for j := 0; j < int(n); j++ {
				next, err := t.reader.Next(ctx)
				if err != nil {
					return nil, err
				}
				docOff += int(next)
			}
		}

		for i := 0; i < numFields; i++ {
			termCount, err := numTerms.Get(skip + i)
			if err != nil {
				return nil, err
			}
			fieldSuffixLengths := make([]int, termCount)
			suffixLengths[i] = fieldSuffixLengths
			for j := 0; j < int(termCount); {
				next, err := t.reader.NextSlices(ctx, int(termCount)-j)
				if err != nil {
					return nil, err
				}
				for _, n := range next {
					fieldSuffixLengths[j] = int(n)
					j++
				}
			}
			fieldLengths[i] = lo.Sum(suffixLengths[i])
			docLen += fieldLengths[i]
		}
		totalLen = docOff + docLen
		for i := skip + numFields; i < totalFields; i++ {
			size, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			for j := 0; j < int(size); j++ {
				num, err := t.reader.Next(ctx)
				if err != nil {
					return nil, err
				}
				totalLen += int(num)
			}
		}
	}

	// term freqs
	termFreqs := make([]int, totalTerms)
	{
		t.reader.Reset(t.vectorsStream, totalTerms)
		for i := 0; i < totalTerms; {
			freq, err := t.reader.Next(ctx)
			if err != nil {
				return nil, err
			}
			termFreqs[i] = int(freq)
			i++
		}
	}

	// total number of positions, offsets and payloads
	totalPositions := 0
	totalOffsets := 0
	totalPayloads := 0
	{
		termIndex := 0
		for i := 0; i < totalFields; i++ {
			f, err := flags.Get(i)
			if err != nil {
				return nil, err
			}
			termCount, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			for j := 0; j < int(termCount); j++ {
				freq := termFreqs[termIndex]
				termIndex++
				if (f & VECTORS_POSITIONS) != 0 {
					totalPositions += freq
				}
				if (f & VECTORS_OFFSETS) != 0 {
					totalOffsets += freq
				}
				if (f & VECTORS_PAYLOADS) != 0 {
					totalPayloads += freq
				}
			}
		}
	}

	positionIndex, err := positionIndex(skip, numFields, numTerms, termFreqs)
	if err != nil {
		return nil, err
	}
	var positions, startOffsets, lengths [][]int
	if totalPositions > 0 {
		positions, err = t.readPositions(ctx, skip, numFields, flags, numTerms, termFreqs, VECTORS_POSITIONS, totalPositions, positionIndex)
	} else {
		positions = make([][]int, numFields)
	}

	if totalOffsets > 0 {
		// average number of chars per term
		charsPerTerm := make([]float32, len(fieldNums))
		for i := range charsPerTerm {
			n, err := t.vectorsStream.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			charsPerTerm[i] = math.Float32frombits(n)
		}
		startOffsets, err = t.readPositions(ctx, skip, numFields, flags, numTerms, termFreqs, VECTORS_OFFSETS, totalOffsets, positionIndex)
		lengths, err = t.readPositions(ctx, skip, numFields, flags, numTerms, termFreqs, VECTORS_OFFSETS, totalOffsets, positionIndex)

		for i := 0; i < numFields; i++ {
			fStartOffsets := startOffsets[i]
			fPositions := positions[i]
			// patch offsets from positions
			if fStartOffsets != nil && fPositions != nil {
				fieldCharsPerTerm := charsPerTerm[fieldNumOffs[i]]
				for j := 0; j < len(startOffsets[i]); j++ {
					fStartOffsets[j] += (int)(fieldCharsPerTerm * float32(fPositions[j]))
				}
			}
			if fStartOffsets != nil {
				fPrefixLengths := prefixLengths[i]
				fSuffixLengths := suffixLengths[i]
				fLengths := lengths[i]
				end, err := numTerms.Get(skip + i)
				if err != nil {
					return nil, err
				}
				for j := 0; j < int(end); j++ {
					// delta-decode start offsets and  patch lengths using term lengths
					termLength := fPrefixLengths[j] + fSuffixLengths[j]
					lengths[i][positionIndex[i][j]] += termLength
					for k := positionIndex[i][j] + 1; k < positionIndex[i][j+1]; k++ {
						fStartOffsets[k] += fStartOffsets[k-1]
						fLengths[k] += termLength
					}
				}
			}
		}
	} else {
		lengths = make([][]int, numFields)
		startOffsets = lengths
	}
	if totalPositions > 0 {
		// delta-decode positions
		for i := 0; i < numFields; i++ {
			fPositions := positions[i]
			fPositionIndex := positionIndex[i]
			if fPositions != nil {
				end, err := numTerms.Get(skip + i)
				if err != nil {
					return nil, err
				}
				for j := 0; j < int(end); j++ {
					// delta-decode start offsets
					for k := fPositionIndex[j] + 1; k < fPositionIndex[j+1]; k++ {
						fPositions[k] += fPositions[k-1]
					}
				}
			}
		}
	}

	// payload lengths
	payloadIndex := make([][]int, numFields)
	totalPayloadLength := uint64(0)
	payloadOff := uint64(0)
	payloadLen := uint64(0)
	if totalPayloads > 0 {
		t.reader.Reset(t.vectorsStream, totalPayloads)
		// skip
		termIndex := 0
		for i := 0; i < skip; i++ {
			f, err := flags.Get(i)
			if err != nil {
				return nil, err
			}
			termCount, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			if (f & uint64(VECTORS_PAYLOADS)) != 0 {
				for j := 0; j < int(termCount); j++ {
					freq := termFreqs[termIndex+j]
					for k := 0; k < freq; k++ {
						l, err := t.reader.Next(ctx)
						if err != nil {
							return nil, err
						}
						payloadOff += l
					}
				}
			}
			termIndex += int(termCount)
		}
		totalPayloadLength = payloadOff
		// read doc payload lengths
		for i := 0; i < numFields; i++ {
			f, err := flags.Get(skip + i)
			if err != nil {
				return nil, err
			}
			termCount, err := numTerms.Get(skip + i)
			if err != nil {
				return nil, err
			}
			if (f & VECTORS_PAYLOADS) != 0 {
				totalFreq := positionIndex[i][termCount]
				payloadIndex[i] = make([]int, totalFreq+1)
				posIdx := 0
				payloadIndex[i][posIdx] = int(payloadLen)
				for j := 0; j < int(termCount); j++ {
					freq := termFreqs[termIndex+j]
					for k := 0; k < freq; k++ {
						payloadLength, err := t.reader.Next(ctx)
						if err != nil {
							return nil, err
						}
						payloadLen += payloadLength
						payloadIndex[i][posIdx+1] = int(payloadLen)
						posIdx++
					}
				}
			}
			termIndex += int(termCount)
		}
		totalPayloadLength += payloadLen
		for i := skip + numFields; i < totalFields; i++ {
			f, err := flags.Get(i)
			if err != nil {
				return nil, err
			}
			termCount, err := numTerms.Get(i)
			if err != nil {
				return nil, err
			}
			if (f & VECTORS_PAYLOADS) != 0 {
				for j := 0; j < int(termCount); j++ {
					freq := termFreqs[termIndex+j]
					for k := 0; k < freq; k++ {
						n, err := t.reader.Next(ctx)
						if err != nil {
							return nil, err
						}
						totalPayloadLength += n
					}
				}
			}
			termIndex += int(termCount)
		}
	}

	// decompress data
	buf := new(bytes.Buffer)
	if err := t.decompressor.Decompress(ctx, t.vectorsStream, buf); err != nil {
		return nil, err
	}
	offset := docOff + int(payloadOff)
	bs := buf.Bytes()
	suffixBytes := store.NewBytesRef(bs)
	if err := suffixBytes.Set(offset, docLen); err != nil {
		return nil, err
	}
	payloadBytes := store.NewBytesRef(bs)
	if err := payloadBytes.Set(suffixBytes.Offset()+docLen, int(payloadLen)); err != nil {
		return nil, err
	}

	fieldFlags := make([]int, numFields)
	for i := range fieldFlags {
		n, err := flags.Get(skip + i)
		if err != nil {
			return nil, err
		}
		fieldFlags[i] = int(n)
	}

	fieldNumTerms := make([]int, numFields)
	for i := range fieldNumTerms {
		n, err := flags.Get(skip + i)
		if err != nil {
			return nil, err
		}
		fieldNumTerms[i] = int(n)
	}

	fieldTermFreqs := make([][]int, numFields)
	{
		termIdx := 0
		for i := 0; i < skip; i++ {
			n, err := numTerms.Get(skip + i)
			if err != nil {
				return nil, err
			}
			termIdx += int(n)
		}

		for i := 0; i < numFields; i++ {
			termCount, err := numTerms.Get(skip + i)
			if err != nil {
				return nil, err
			}
			fieldTermFreqs[i] = make([]int, termCount)
			for j := range fieldTermFreqs[i] {
				fieldTermFreqs[i][j] = termFreqs[termIdx]
				termIdx++
			}
		}
	}

	return t.newTVFields(fieldNums, fieldFlags, fieldNumOffs, fieldNumTerms, fieldLengths,
		prefixLengths, suffixLengths, fieldTermFreqs,
		positionIndex, positions, startOffsets, lengths,
		payloadBytes, payloadIndex,
		suffixBytes), nil
}

// field -> term index -> position index
func positionIndex(skip, numFields int, numTerms packed.Reader, termFreqs []int) ([][]int, error) {
	positionIndex := make([][]int, numFields)

	termIndex := 0
	for i := 0; i < skip; i++ {
		termCount, err := numTerms.Get(i)
		if err != nil {
			return nil, err
		}
		termIndex += int(termCount)
	}
	for i := 0; i < numFields; i++ {
		termCount, err := numTerms.Get(skip + i)
		if err != nil {
			return nil, err
		}
		positionIndex[i] = make([]int, termCount+1)
		for j := 0; j < int(termCount); j++ {
			freq := termFreqs[termIndex+j]
			positionIndex[i][j+1] = positionIndex[i][j] + freq
		}
		termIndex += int(termCount)
	}
	return positionIndex, nil
}

func (t *TermVectorsReader) readPositions(ctx context.Context, skip, numFields int, flags, numTerms packed.Reader,
	termFreqs []int, flag, totalPositions int, positionIndex [][]int) ([][]int, error) {
	positions := make([][]int, numFields)
	t.reader.Reset(t.vectorsStream, totalPositions)
	// skip
	toSkip := 0
	termIndex := 0
	for i := 0; i < skip; i++ {
		f, err := flags.Get(i)
		if err != nil {
			return nil, err
		}
		termCount, err := numTerms.Get(i)
		if err != nil {
			return nil, err
		}
		if (f & uint64(flag)) != 0 {
			for j := 0; j < int(termCount); j++ {
				freq := termFreqs[termIndex+j]
				toSkip += freq
			}
		}
		termIndex += int(termCount)
	}

	if err := t.reader.Skip(ctx, toSkip); err != nil {
		return nil, err
	}
	// read doc positions
	for i := 0; i < numFields; i++ {
		f, err := flags.Get(i)
		if err != nil {
			return nil, err
		}
		termCount, err := numTerms.Get(i)
		if err != nil {
			return nil, err
		}
		if (f & uint64(flag)) != 0 {
			totalFreq := positionIndex[i][termCount]
			fieldPositions := make([]int, totalFreq)
			positions[i] = fieldPositions
			for j := 0; j < totalFreq; j++ {
				position, err := t.reader.Next(ctx)
				if err != nil {
					return nil, err
				}
				fieldPositions[j] = int(position)

			}
		}
		termIndex += int(termCount)
	}
	err := t.reader.Skip(ctx, totalPositions-t.reader.Ord())
	if err != nil {
		return nil, err
	}

	return positions, nil
}

func (t *TermVectorsReader) CheckIntegrity() error {
	if err := t.indexReader.CheckIntegrity(); err != nil {
		return err
	}
	_, err := codecs.ChecksumEntireFile(context.Background(), t.vectorsStream)
	return err
}

func (t *TermVectorsReader) Clone(ctx context.Context) index.TermVectorsReader {
	reader, _ := _newTermVectorsReader(t)
	return reader
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

	fieldNums     []int
	fieldFlags    []int
	fieldNumOffs  []int
	numTerms      []int
	fieldLengths  []int
	prefixLengths [][]int
	suffixLengths [][]int
	termFreqs     [][]int
	positionIndex [][]int
	positions     [][]int
	startOffsets  [][]int
	lengths       [][]int
	payloadIndex  [][]int
	suffixBytes   *store.BytesRef
	payloadBytes  *store.BytesRef
}

func (t *TermVectorsReader) NewTVFields() *TVFields {
	return nil
}

func (t *TermVectorsReader) newTVFields(fieldNums, fieldFlags, fieldNumOffs, numTerms, fieldLengths []int,
	prefixLengths, suffixLengths, termFreqs [][]int,
	positionIndex, positions, startOffsets, lengths [][]int,
	payloadBytes *store.BytesRef, payloadIndex [][]int,
	suffixBytes *store.BytesRef) *TVFields {

	return &TVFields{
		TermVectorsReader: t,
		fieldNums:         fieldNums,
		fieldFlags:        fieldFlags,
		fieldNumOffs:      fieldNumOffs,
		numTerms:          numTerms,
		fieldLengths:      fieldLengths,
		prefixLengths:     prefixLengths,
		suffixLengths:     suffixLengths,
		termFreqs:         termFreqs,
		positionIndex:     positionIndex,
		positions:         positions,
		startOffsets:      startOffsets,
		lengths:           lengths,
		payloadIndex:      payloadIndex,
		suffixBytes:       suffixBytes,
		payloadBytes:      payloadBytes,
	}
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
	fieldLen := -1
	for i := 0; i < len(f.fieldNumOffs); i++ {
		if i < idx {
			fieldOff += f.fieldLengths[i]
		} else {
			fieldLen = f.fieldLengths[i]
			break
		}
	}

	return NewTVTerms(f.numTerms[idx], f.fieldFlags[idx],
		f.prefixLengths[idx], f.suffixLengths[idx], f.termFreqs[idx],
		f.positionIndex[idx], f.positions[idx], f.startOffsets[idx], f.lengths[idx],
		f.payloadIndex[idx], f.payloadBytes,
		store.NewMustBytesRef(f.suffixBytes.RawBytes(), f.suffixBytes.Offset()+fieldOff, fieldLen)), nil
}

func (f *TVFields) Size() int {
	return len(f.fieldNumOffs)
}

var _ index.Terms = &TVTerms{}

type TVTerms struct {
	*coreIndex.BaseTerms

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
	termBytes     *store.BytesRef
	payloadBytes  *store.BytesRef
}

func NewTVTerms(numTerms, flags int,
	prefixLengths, suffixLengths, termFreqs, positionIndex []int,
	positions, startOffsets, lengths, payloadIndex []int,
	termBytes, payloadBytes *store.BytesRef) *TVTerms {

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
	termsEnum := NewTVTermsEnum()
	if err := termsEnum.reset(t.numTerms, t.flags, t.prefixLengths, t.suffixLengths, t.termFreqs,
		t.positionIndex, t.positions, t.startOffsets, t.lengths, t.payloadIndex, t.payloadBytes,
		store.NewByteArrayDataInput(t.termBytes.Bytes())); err != nil {
		return nil, err
	}
	return termsEnum, nil
}

func (t *TVTerms) Size() (int, error) {
	return t.numTerms, nil
}

func (t *TVTerms) GetSumTotalTermFreq() (int64, error) {
	return int64(t.totalTermFreq), nil
}

func (t *TVTerms) GetSumDocFreq() (int64, error) {
	return int64(t.numTerms), nil
}

func (t *TVTerms) GetDocCount() (int, error) {
	return 1, nil
}

func (t *TVTerms) HasFreqs() bool {
	return true
}

func (t *TVTerms) HasOffsets() bool {
	return (t.flags & VECTORS_OFFSETS) != 0
}

func (t *TVTerms) HasPositions() bool {
	return (t.flags & VECTORS_POSITIONS) != 0
}

func (t *TVTerms) HasPayloads() bool {
	return (t.flags & VECTORS_PAYLOADS) != 0
}

var _ index.TermsEnum = &TVTermsEnum{}

type TVTermsEnum struct {
	*coreIndex.BaseTermsEnum

	numTerms      int
	startPos      int64
	ord           int
	prefixLengths []int
	suffixLengths []int
	termFreqs     []int
	positionIndex []int
	positions     []int
	startOffsets  []int
	lengths       []int
	payloadIndex  []int
	in            *store.ByteArrayDataInput
	payloads      *store.BytesRef
	term          []byte
}

func NewTVTermsEnum() *TVTermsEnum {
	return &TVTermsEnum{
		// TODO：reuse buffer
		term: make([]byte, 0),
	}
}

func (t *TVTermsEnum) reset(numTerms, _flags int,
	prefixLengths, suffixLengths, termFreqs, positionIndex []int,
	positions, startOffsets, lengths, payloadIndex []int,
	payloads *store.BytesRef, in *store.ByteArrayDataInput) error {

	t.numTerms = numTerms
	t.prefixLengths = prefixLengths
	t.suffixLengths = suffixLengths
	t.termFreqs = termFreqs
	t.positionIndex = positionIndex
	t.positions = positions
	t.startOffsets = startOffsets
	t.lengths = lengths
	t.payloadIndex = payloadIndex
	t.payloads = payloads
	t.in = in
	//t.startPos = in.GetPosition()
	return t._reset()
}

func (t *TVTermsEnum) _reset() error {
	t.term = t.term[:0]
	if _, err := t.in.Seek(t.startPos, io.SeekStart); err != nil {
		return err
	}
	t.ord = -1
	return nil
}

func (t *TVTermsEnum) Next(ctx context.Context) ([]byte, error) {
	if t.ord == t.numTerms-1 {
		return nil, nil
	}

	t.ord++
	size := t.prefixLengths[t.ord] + t.suffixLengths[t.ord]
	slices.Grow(t.term, size)
	t.term = t.term[:size]
	if _, err := t.in.Read(t.term); err != nil {
		return nil, err
	}
	return t.term, nil
}

func (t *TVTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	if t.ord < t.numTerms && t.ord >= 0 {
		termBytes, err := t.Term()
		if err != nil {
			return 0, err
		}
		cmp := bytes.Compare(termBytes, text)
		if cmp == 0 {
			return index.SEEK_STATUS_FOUND, nil
		}
		if cmp > 0 {
			if err := t._reset(); err != nil {
				return 0, err
			}
		}
	}

	// linear scan
	for {
		termBytes, err := t.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return index.SEEK_STATUS_END, nil
			}
			return 0, err
		}

		cmp := bytes.Compare(termBytes, text)
		if cmp > 0 {
			return index.SEEK_STATUS_NOT_FOUND, nil
		}

		if cmp == 0 {
			return index.SEEK_STATUS_FOUND, nil
		}
	}
}

func (t *TVTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	return errors.New("unsupported operation exception")
}

func (t *TVTermsEnum) Term() ([]byte, error) {
	return t.term, nil
}

func (t *TVTermsEnum) Ord() (int64, error) {
	return 0, errors.New("unsupported operation")
}

func (t *TVTermsEnum) DocFreq() (int, error) {
	return 1, nil
}

func (t *TVTermsEnum) TotalTermFreq() (int64, error) {
	return int64(t.termFreqs[t.ord]), nil
}

func (t *TVTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	var docsEnum *TVPostingsEnum
	if reuse == nil {
		docsEnum = NewTVPostingsEnum()
	} else {
		if enum, ok := reuse.(*TVPostingsEnum); ok {
			docsEnum = enum
		} else {
			docsEnum = NewTVPostingsEnum()
		}
	}
	ord := t.ord
	docsEnum.reset(t.termFreqs[ord], t.positionIndex[ord],
		t.positions, t.startOffsets, t.lengths, t.payloads, t.payloadIndex)
	return docsEnum, nil
}

func (t *TVTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	delegate, err := t.Postings(nil, coreIndex.POSTINGS_ENUM_FREQS)
	if err != nil {
		return nil, err
	}
	return coreIndex.NewSlowImpactsEnum(delegate), nil
}

var _ index.PostingsEnum = &TVPostingsEnum{}

type TVPostingsEnum struct {
	doc           int
	termFreq      int
	positionIndex int
	positions     []int
	startOffsets  []int
	lengths       []int
	payload       *store.BytesRef
	payloadIndex  []int
	i             int
	isEOF         bool
}

func NewTVPostingsEnum() *TVPostingsEnum {
	return &TVPostingsEnum{
		doc:     -1,
		payload: new(store.BytesRef),
	}
}

func (t *TVPostingsEnum) checkDoc() error {
	if t.doc == types.NO_MORE_DOCS {
		return errors.New("docsEnum exhausted")
	}
	if t.doc < 0 {
		return errors.New("docsEnum not started")
	}
	return nil
}

func (t *TVPostingsEnum) checkPosition() error {
	if err := t.checkDoc(); err != nil {
		return err
	}
	if t.i < 0 {
		return errors.New("position enum not started")
	}
	if t.i >= t.termFreq {
		return errors.New("read past last position")
	}
	return nil
}

func (t *TVPostingsEnum) DocID() int {
	return t.doc
}

func (t *TVPostingsEnum) NextDoc(ctx context.Context) (int, error) {
	if t.doc == -1 {
		t.doc = 0
		return 0, nil
	} else {
		t.doc = types.NO_MORE_DOCS
		return 0, io.EOF
	}
}

func (t *TVPostingsEnum) Advance(ctx context.Context, target int) (int, error) {
	return t.SlowAdvance(ctx, target)
}

func (t *TVPostingsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, t, target)
}

func (t *TVPostingsEnum) Cost() int64 {
	return 1
}

func (t *TVPostingsEnum) Freq() (int, error) {
	if err := t.checkDoc(); err != nil {
		return 0, err
	}
	return t.termFreq, nil
}

func (t *TVPostingsEnum) NextPosition() (int, error) {
	if t.doc != 0 {
		return 0, errors.New("illegal state exception")
	}
	if t.i >= t.termFreq-1 {
		return 0, errors.New("read past last position")
	}

	t.i++

	if len(t.payloadIndex) != 0 {
		offset := t.payloadIndex[t.positionIndex+t.i]
		length := t.payloadIndex[t.positionIndex+t.i+1] - t.payloadIndex[t.positionIndex+t.i]
		if err := t.payload.Set(offset, length); err != nil {
			return 0, err
		}
	}

	if len(t.positions) == 0 {
		return -1, nil
	} else {
		return t.positions[t.positionIndex+t.i], nil
	}
}

func (t *TVPostingsEnum) StartOffset() (int, error) {
	if err := t.checkPosition(); err != nil {
		return 0, err
	}
	if len(t.startOffsets) == 0 {
		return -1, nil
	}
	return t.startOffsets[t.positionIndex+t.i], nil
}

func (t *TVPostingsEnum) EndOffset() (int, error) {
	if err := t.checkPosition(); err != nil {
		return 0, err
	}
	if t.startOffsets == nil {
		return -1, nil
	}
	return t.startOffsets[t.positionIndex+t.i] + t.lengths[t.positionIndex+t.i], nil
}

func (t *TVPostingsEnum) GetPayload() ([]byte, error) {
	if err := t.checkPosition(); err != nil {
		return nil, err
	}

	if t.payloadIndex == nil || t.payload.Len() == 0 {
		return []byte{}, nil
	}
	return t.payload.Bytes(), nil
}

func (t *TVPostingsEnum) reset(freq, positionIndex int,
	positions, startOffsets, lengths []int,
	payloads *store.BytesRef, payloadIndex []int) {

	t.termFreq = freq
	t.positionIndex = positionIndex
	t.positions = positions
	t.startOffsets = startOffsets
	t.lengths = lengths
	t.payload = payloads
	t.payloadIndex = payloadIndex

	t.doc = -1
	t.i = -1
}
