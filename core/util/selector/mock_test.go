package selector

import "bytes"

var (
	_ RadixSelector = &MockRadix{}
	_ IntroSelector = &MockRadix{}
)

type MockRadix struct {
	values [][]byte
}

func NewMockRadix(values ...[]byte) *MockRadix {
	radix := &MockRadix{
		values: make([][]byte, 0),
	}
	radix.values = append(radix.values, values...)
	return radix
}

func (m *MockRadix) Add(bs []byte) {
	m.values = append(m.values, bs)
}

func (m *MockRadix) Swap(i, j int) {
	m.values[i], m.values[j] = m.values[j], m.values[i]
}

func (m *MockRadix) ByteAt(i int, k int) int {
	if i >= len(m.values) {
		return -1
	}

	if k >= len(m.values[i]) {
		return -1
	}

	b := m.values[i][k]
	return int(b)
}

func (m *MockRadix) Value(i int) []byte {
	return m.values[i]
}

func (m *MockRadix) Compare(i, j int) int {
	return bytes.Compare(m.values[i], m.values[j])
}
