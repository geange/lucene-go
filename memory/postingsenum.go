package memory

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"io"

	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

var _ index.PostingsEnum = &memPostingsEnum{}

type memPostingsEnum struct {
	sliceReader       *ints.SliceReader
	posUpto           int
	hasNext           bool
	doc               int
	freq              int
	pos               int
	startOffset       int
	endOffset         int
	payloadIndex      int
	payloadBuilder    *bytesref.Builder
	payloadsBytesRefs *bytesref.Array
	storeOffsets      bool
	storePayloads     bool
}

func newPostingsEnum(intBlockPool *ints.BlockPool, storePayloads bool) *memPostingsEnum {
	return &memPostingsEnum{
		doc:         -1,
		sliceReader: ints.NewSliceReader(intBlockPool),
		payloadBuilder: func() *bytesref.Builder {
			if storePayloads {
				return bytesref.NewBytesRefBuilder()
			}
			return nil
		}(),
	}
}

func (m *memPostingsEnum) reset(start, end, freq int) index.PostingsEnum {
	m.sliceReader.Reset(start, end)
	m.posUpto = 0
	m.hasNext = true
	m.doc = -1
	m.freq = freq
	return m
}

func (m *memPostingsEnum) DocID() int {
	return m.doc
}

func (m *memPostingsEnum) NextDoc() (int, error) {
	m.pos = -1
	if m.hasNext {
		m.hasNext = false
		m.doc = 0
		return m.doc, nil
	} else {
		return 0, io.EOF
	}
}

func (m *memPostingsEnum) Advance(ctx context.Context, target int) (int, error) {
	return m.SlowAdvance(nil, target)
}

func (m *memPostingsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = m.NextDoc()
		if err != nil {
			return 0, nil
		}
	}
	return doc, nil
}

func (m *memPostingsEnum) Cost() int64 {
	return 1
}

func (m *memPostingsEnum) Freq() (int, error) {
	return m.freq, nil
}

func (m *memPostingsEnum) NextPosition() (int, error) {
	m.posUpto++
	pos := m.sliceReader.ReadInt()
	if m.storeOffsets {
		//pos = sliceReader.readInt();
		m.startOffset = m.sliceReader.ReadInt()
		m.endOffset = m.sliceReader.ReadInt()
	}
	if m.storePayloads {
		m.payloadIndex = m.sliceReader.ReadInt()
	}
	return pos, nil
}

func (m *memPostingsEnum) StartOffset() (int, error) {
	return m.startOffset, nil
}

func (m *memPostingsEnum) EndOffset() (int, error) {
	return m.endOffset, nil
}

func (m *memPostingsEnum) GetPayload() ([]byte, error) {
	if m.payloadIndex == -1 {
		return nil, nil
	}
	return m.payloadsBytesRefs.Get(m.payloadBuilder, m.payloadIndex), nil
}
