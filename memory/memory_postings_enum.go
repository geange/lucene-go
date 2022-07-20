package memory

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var _ index.PostingsEnum = &MemoryPostingsEnum{}

type MemoryPostingsEnum struct {
	sliceReader    *util.SliceReader
	posUpto        int
	hasNext        bool
	doc            int
	freq           int
	pos            int
	startOffset    int
	endOffset      int
	payloadIndex   int
	payloadBuilder *util.BytesRefBuilder

	payloadsBytesRefs *util.BytesRefArray
	storeOffsets      bool
	storePayloads     bool
}

func NewMemoryPostingsEnum() *MemoryPostingsEnum {
	return &MemoryPostingsEnum{}
}

func (m *MemoryPostingsEnum) reset(start, end, freq int) index.PostingsEnum {
	m.sliceReader.Reset(start, end)
	m.posUpto = 0
	m.hasNext = true
	m.doc = -1
	m.freq = freq
	return m
}

func (m *MemoryPostingsEnum) DocID() int {
	return m.doc
}

func (m *MemoryPostingsEnum) NextDoc() (int, error) {
	m.pos = -1
	if m.hasNext {
		m.hasNext = false
		m.doc = 0
		return m.doc, nil
	} else {
		m.doc = math.MaxInt32
		return m.doc, nil
	}
}

func (m *MemoryPostingsEnum) Advance(target int) (int, error) {
	return m.SlowAdvance(target)
}

func (m *MemoryPostingsEnum) SlowAdvance(target int) (int, error) {
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

func (m *MemoryPostingsEnum) Cost() int64 {
	return 1
}

func (m *MemoryPostingsEnum) Freq() (int, error) {
	return m.freq, nil
}

func (m *MemoryPostingsEnum) NextPosition() (int, error) {
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

func (m *MemoryPostingsEnum) StartOffset() (int, error) {
	return m.startOffset, nil
}

func (m *MemoryPostingsEnum) EndOffset() (int, error) {
	return m.endOffset, nil
}

func (m *MemoryPostingsEnum) GetPayload() ([]byte, error) {
	if m.payloadIndex == -1 {
		return nil, nil
	}
	return m.payloadsBytesRefs.Get(m.payloadBuilder, m.payloadIndex), nil
}
