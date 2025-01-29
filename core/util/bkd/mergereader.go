package bkd

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/types"
)

type MergeReader struct {
	bkd    *Reader
	state  *IntersectState
	docMap types.DocMap

	// Current doc ID
	docID int

	// Which doc in this block we are up to
	docBlockUpto int

	// How many docs in the current block
	docsInBlock int

	// Which leaf block we are up to
	blockID int

	packedValues []byte
}

func (m *MergeReader) Next(ctx context.Context) (bool, error) {
	for {
		if m.docBlockUpto == m.docsInBlock {
			if m.blockID == m.bkd.leafNodeOffset {
				//System.out.println("  done!");
				return false, nil
			}
			//System.out.println("  new block @ fp=" + state.in.getFilePointer());
			var err error
			m.docsInBlock, err = m.bkd.readDocIDs(ctx, m.state.in, m.state.in.GetFilePointer(), m.state.scratchIterator)
			if err != nil {
				return false, err
			}
			//assert docsInBlock > 0;
			m.docBlockUpto = 0

			err = m.bkd.visitDocValues(ctx, m.state.commonPrefixLengths, m.state.scratchDataPackedValue, m.state.scratchMinIndexPackedValue, m.state.scratchMaxIndexPackedValue, m.state.in, m.state.scratchIterator, m.docsInBlock, &mergeReaderVisitor{i: 0, p: m})
			if err != nil {
				return false, err
			}

			m.blockID++
		}

		index := m.docBlockUpto
		m.docBlockUpto++
		oldDocID := m.state.scratchIterator.docIDs[index]

		var mappedDocID int
		if m.docMap == nil {
			mappedDocID = oldDocID
		} else {
			mappedDocID = m.docMap.Get(oldDocID)
		}

		if mappedDocID != -1 {
			// Not deleted!
			m.docID = mappedDocID

			srcFrom := index * m.bkd.config.packedBytesLength
			srcTo := srcFrom + m.bkd.config.packedBytesLength

			destFrom := 0
			destTo := destFrom + m.bkd.config.packedBytesLength

			copy(m.state.scratchDataPackedValue[destFrom:destTo], m.packedValues[srcFrom:srcTo])
			return true, nil
		}
	}
}

var _ types.IntersectVisitor = &mergeReaderVisitor{}

type mergeReaderVisitor struct {
	i int
	p *MergeReader
}

func (m *mergeReaderVisitor) Visit(ctx context.Context, docID int) error {
	return errors.New("UnsupportedOperationException")
}

func (m *mergeReaderVisitor) VisitLeaf(ctx context.Context, docID int, packedValue []byte) error {
	bkd := m.p.bkd
	arraycopy(packedValue, 0, m.p.packedValues, m.i*bkd.config.packedBytesLength, bkd.config.packedBytesLength)
	m.i++
	return nil
}

func (m *mergeReaderVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	return types.CELL_CROSSES_QUERY
}

func (m *mergeReaderVisitor) Grow(count int) {
}
