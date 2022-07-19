package memory

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
)

var _ index.PostingsEnum = &MemoryPostingsEnum{}

type MemoryPostingsEnum struct {
}

func (m *MemoryPostingsEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryPostingsEnum) GetPayload() (*util.BytesRef, error) {
	//TODO implement me
	panic("implement me")
}
