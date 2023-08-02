package sorter

import (
	"bytes"
	"sort"
)

var (
	_ MSBRadixInterface = &Mock{}
	_ sort.Interface    = &MockInt{}
)

type Mock struct {
	values [][]byte
}

func (m *Mock) Len() int {
	return len(m.values)
}

func (m *Mock) Less(i, j int) bool {
	return bytes.Compare(m.values[i], m.Value(j)) < 0
}

func (m *Mock) ByteAt(i int, k int) int {
	if i >= len(m.values) {
		return -1
	}

	if k >= len(m.values[i]) {
		return -1
	}

	b := m.values[i][k]
	return int(b)
}

func (m *Mock) Value(i int) []byte {
	return m.values[i]
}

func NewMock(values ...[]byte) *Mock {
	radix := &Mock{
		values: make([][]byte, 0),
	}
	radix.values = append(radix.values, values...)
	return radix
}

func (m *Mock) Add(bs []byte) {
	m.values = append(m.values, bs)
}

func (m *Mock) Swap(i, j int) {
	if i >= m.Len() || j >= m.Len() {
		panic("over size")
	}
	m.values[i], m.values[j] = m.values[j], m.values[i]
}

func (m *Mock) Compare(i, j, skipBytes int) int {
	return bytes.Compare(m.values[i][skipBytes:], m.values[j][skipBytes:])
}

//func (m *Mock) Compare(i, j int) int {
//	return bytes.Compare(m.values[i], m.values[j])
//}

type MockInt struct {
	values []int
}

func (m *MockInt) Len() int {
	return len(m.values)
}

func (m *MockInt) Less(i, j int) bool {
	return m.values[i] < m.values[j]
}

func NewMockInt(values ...int) *MockInt {
	radix := &MockInt{
		values: make([]int, 0),
	}
	radix.values = append(radix.values, values...)
	return radix
}

func (m *MockInt) Add(v int) {
	m.values = append(m.values, v)
}

func (m *MockInt) Swap(i, j int) {
	m.values[i], m.values[j] = m.values[j], m.values[i]
}

func (m *MockInt) Compare(i, j int) int {
	if m.values[i] < m.values[j] {
		return -1
	}
	if m.values[i] > m.values[j] {
		return 1
	}
	return 0
}
