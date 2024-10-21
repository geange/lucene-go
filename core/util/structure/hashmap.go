package structure

type Hash interface {
	Hash() int64
}

type Map[K Hash, V any] struct {
	mp     map[int64]int
	values []*MapEntry[K, V]
	rmIdx  []int
}

func NewMap[K Hash, V any]() *Map[K, V] {
	return &Map[K, V]{
		mp:     map[int64]int{},
		values: make([]*MapEntry[K, V], 0),
		rmIdx:  make([]int, 0),
	}
}

func (m *Map[K, V]) Put(key K, value V) {
	code := key.Hash()
	idx, ok := m.mp[code]
	if ok {
		m.values[idx].Key = key
		m.values[idx].Value = value
		return
	}
	if len(m.rmIdx) == 0 {
		m.values = append(m.values, &MapEntry[K, V]{Key: key, Value: value})
		m.mp[code] = len(m.values) - 1
		return
	}
	idx = m.rmIdx[len(m.rmIdx)-1]  
	m.mp[code] = idx
	m.values[idx].Key = key
	m.values[idx].Value = value
	m.rmIdx = m.rmIdx[:len(m.rmIdx)-1]
}

func (m *Map[K, V]) Get(key K) (v V, ok bool) {
	code := key.Hash()
	idx, ok := m.mp[code]
	if ok {
		return m.values[idx].Value, true
	}
	return
}

func (m *Map[K, V]) Remove(key K) bool {
	code := key.Hash()
	idx, ok := m.mp[code]
	if !ok {
		return false
	}

	delete(m.mp, code)
	m.rmIdx = append(m.rmIdx, idx)
	return true
}

func (m *Map[K, V]) Clear() {
	m.values = m.values[:0]
	m.rmIdx = m.rmIdx[:0]
	clear(m.mp)
}

type MapEntry[K Hash, V any] struct {
	Key   K
	Value V
}
