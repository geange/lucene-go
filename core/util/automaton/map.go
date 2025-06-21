package automaton

import (
	"iter"
	"sync"
)

// Hashable 自定义哈希接口
type Hashable interface {
	Hash() uint64
	Equals(other Hashable) bool
}

// HashMap 自定义哈希表结构
type HashMap[T any] struct {
	buckets     []*Entry[T]
	size        int
	mask        uint64
	mutex       sync.RWMutex // 可选并发控制
	emptyValue  T
	loadFactory float64
}

// Entry 哈希表条目
type Entry[T any] struct {
	key   Hashable
	value T
	next  *Entry[T]
}

type optionsHashMap struct {
	capacity    int     // 默认4
	loadFactory float64 // 负载因子，默认0.75
}

func newOptionsHashMap(opts ...OptionsHashMap) *optionsHashMap {
	options := &optionsHashMap{
		capacity:    1,
		loadFactory: 0.75,
	}

	for _, opt := range opts {
		opt(options)
	}

	realCap := 1
	for realCap < options.capacity {
		realCap <<= 1
	}
	options.capacity = realCap

	return options
}

type OptionsHashMap func(hashMap *optionsHashMap)

func WithCapacity(capacity int) OptionsHashMap {
	return func(hashMap *optionsHashMap) {
		hashMap.capacity = capacity
	}
}

func WithLoadFactory(loadFactory float64) OptionsHashMap {
	return func(hashMap *optionsHashMap) {
		hashMap.loadFactory = loadFactory
	}
}

// NewHashMap 创建哈希表
// 参数：capacity 初始容量（自动调整为2的幂）
func NewHashMap[T any](options ...OptionsHashMap) *HashMap[T] {
	opt := newOptionsHashMap(options...)

	return &HashMap[T]{
		buckets:     make([]*Entry[T], opt.capacity),
		mask:        uint64(opt.capacity - 1),
		loadFactory: opt.loadFactory,
	}
}

// Set 插入键值对
func (m *HashMap[T]) Set(key Hashable, value T) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	hash := key.Hash()
	index := hash & m.mask

	// 遍历链表查找是否存在相同key
	for e := m.buckets[index]; e != nil; e = e.next {
		if e.key.Equals(key) {
			e.value = value // 更新已有值
			return
		}
	}

	// 头插法添加新条目
	m.buckets[index] = &Entry[T]{
		key:   key,
		value: value,
		next:  m.buckets[index],
	}
	m.size++

	// 自动扩容（当负载因子>0.75时）
	if float64(m.size)/float64(len(m.buckets)) > m.loadFactory {
		m.resize()
	}
}

// Get 获取值
func (m *HashMap[T]) Get(key Hashable) (T, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	hash := key.Hash()
	index := hash & m.mask

	for e := m.buckets[index]; e != nil; e = e.next {
		if e.key.Equals(key) {
			return e.value, true
		}
	}
	return m.emptyValue, false
}

// Delete 删除键
func (m *HashMap[T]) Delete(key Hashable) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	hash := key.Hash()
	index := hash & m.mask

	var prev *Entry[T]
	for e := m.buckets[index]; e != nil; prev, e = e, e.next {
		if e.key.Equals(key) {
			if prev == nil {
				m.buckets[index] = e.next
			} else {
				prev.next = e.next
			}
			m.size--
			return
		}
	}
}

// 扩容哈希表
func (m *HashMap[T]) resize() {
	newCap := len(m.buckets) << 1
	newBuckets := make([]*Entry[T], newCap)
	newMask := uint64(newCap - 1)

	// 重新哈希所有条目
	for _, head := range m.buckets {
		for e := head; e != nil; e = e.next {
			newIndex := e.key.Hash() & newMask
			newBuckets[newIndex] = &Entry[T]{
				key:   e.key,
				value: e.value,
				next:  newBuckets[newIndex],
			}
		}
	}

	m.buckets = newBuckets
	m.mask = newMask
}

// Size 获取元素数量
func (m *HashMap[T]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.size
}

func (m *HashMap[T]) Iterator() iter.Seq2[Hashable, T] {
	return func(yield func(Hashable, T) bool) {
		for _, bucket := range m.buckets {
			if bucket == nil {
				continue
			}
			for e := bucket; e != nil; e = e.next {
				if !yield(e.key, e.value) {
					return
				}
			}
		}
	}
}
