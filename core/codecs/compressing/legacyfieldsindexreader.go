package compressing

import (
	"context"
	"fmt"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
	"github.com/geange/lucene-go/core/util/zigzag"
)

var _ FieldsIndex = &LegacyFieldsIndexReader{}

type LegacyFieldsIndexReader struct {
	maxDoc              int
	docBases            []int
	startPointers       []int64
	avgChunkDocs        []int
	avgChunkSizes       []int
	docBasesDeltas      []packed.Reader // delta from the avg
	startPointersDeltas []packed.Reader // delta from the avg
}

// NewLegacyFieldsIndexReader
// It is the responsibility of the caller to close fieldsIndexIn after this constructor
// has been called
func NewLegacyFieldsIndexReader(ctx context.Context, fieldsIndexIn store.IndexInput, si index.SegmentInfo) (*LegacyFieldsIndexReader, error) {
	maxDoc, err := si.MaxDoc()
	if err != nil {
		return nil, err
	}

	docBases := make([]int, 0, 16)
	startPointers := make([]int64, 0, 16)
	avgChunkDocs := make([]int, 0, 16)
	avgChunkSizes := make([]int, 0, 16)
	docBasesDeltas := make([]packed.Reader, 0, 16)
	startPointersDeltas := make([]packed.Reader, 0, 16)
	packedIntsVersion, err := fieldsIndexIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	for {
		numChunks, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if numChunks == 0 {
			break
		}

		// doc bases
		docBase, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		docBases = append(docBases, int(docBase))
		avgChunkDoc, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		avgChunkDocs = append(avgChunkDocs, int(avgChunkDoc))
		bitsPerDocBase, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if bitsPerDocBase > 32 {
			return nil, fmt.Errorf("corrupted bitsPerDocBase: %d", bitsPerDocBase)
		}
		docBasesDelta, err := packed.GetReaderNoHeader(ctx, fieldsIndexIn, packed.FormatPacked,
			int(packedIntsVersion), int(numChunks), int(bitsPerDocBase))
		if err != nil {
			return nil, err
		}
		docBasesDeltas = append(docBasesDeltas, docBasesDelta)

		// start pointers
		startPointer, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		startPointers = append(startPointers, int64(startPointer))
		avgChunkSize, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		avgChunkSizes = append(avgChunkSizes, int(avgChunkSize))
		bitsPerStartPointer, err := fieldsIndexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if bitsPerStartPointer > 64 {
			return nil, fmt.Errorf("corrupted bitsPerStartPointer: %d", bitsPerStartPointer)
		}
		startPointersDelta, err := packed.GetReaderNoHeader(ctx, fieldsIndexIn, packed.FormatPacked,
			int(packedIntsVersion), int(numChunks), int(bitsPerStartPointer))
		if err != nil {
			return nil, err
		}
		startPointersDeltas = append(startPointersDeltas, startPointersDelta)
	}

	return &LegacyFieldsIndexReader{
		maxDoc:              maxDoc,
		docBases:            docBases,
		startPointers:       startPointers,
		avgChunkDocs:        avgChunkDocs,
		avgChunkSizes:       avgChunkSizes,
		docBasesDeltas:      docBasesDeltas,
		startPointersDeltas: docBasesDeltas,
	}, nil
}

func (r *LegacyFieldsIndexReader) Close() error {
	return nil
}

func (r *LegacyFieldsIndexReader) GetStartPointer(docID int) (int64, error) {
	block := r.block(docID)
	relativeChunk, err := r.relativeChunk(block, docID-r.docBases[block])
	if err != nil {
		return 0, err
	}
	startPointer, err := r.relativeStartPointer(block, relativeChunk)
	if err != nil {
		return 0, err
	}
	return r.startPointers[block] + int64(startPointer), nil
}

func (r *LegacyFieldsIndexReader) block(docID int) int {
	lo := 0
	hi := len(r.docBases) - 1
	for lo <= hi {
		mid := (lo + hi) >> 1
		midValue := r.docBases[mid]
		if midValue == docID {
			return mid
		} else if midValue < docID {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return hi
}

func (r *LegacyFieldsIndexReader) relativeDocBase(block, relativeChunk int) (int, error) {
	expected := r.avgChunkDocs[block] * relativeChunk

	idx, err := r.docBasesDeltas[block].Get(relativeChunk)
	if err != nil {
		return 0, err
	}
	delta := zigzag.Decode(idx)
	return expected + int(delta), nil
}

func (r *LegacyFieldsIndexReader) relativeStartPointer(block, relativeChunk int) (int, error) {
	expected := r.avgChunkSizes[block] * relativeChunk
	idx, err := r.startPointersDeltas[block].Get(relativeChunk)
	if err != nil {
		return 0, err
	}

	delta := int(zigzag.Decode(idx))
	return expected + delta, nil
}

func (r *LegacyFieldsIndexReader) relativeChunk(block, relativeDoc int) (int, error) {
	lo := 0
	hi := r.docBasesDeltas[block].Size() - 1
	for lo <= hi {
		mid := (lo + hi) >> 1
		midValue, err := r.relativeDocBase(block, mid)
		if err != nil {
			return 0, err
		}
		if midValue == relativeDoc {
			return mid, nil
		} else if midValue < relativeDoc {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return hi, nil
}

func (r *LegacyFieldsIndexReader) CheckIntegrity() error {
	return nil
}

func (r *LegacyFieldsIndexReader) Clone() (FieldsIndex, error) {
	return r, nil
}
