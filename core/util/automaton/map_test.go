package automaton

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试键类型
type TestKey struct {
	part1 int
	part2 string
}

func (k TestKey) Hash() uint64 {
	return uint64(k.part1 + len(k.part2))
}

func (k TestKey) Equals(other Hashable) bool {
	o, ok := other.(TestKey)
	return ok && k.part1 == o.part1 && k.part2 == o.part2
}

// 另一个测试键类型（用于类型安全测试）
type AnotherKey int

func (k AnotherKey) Hash() uint64 {
	return uint64(k)
}

func (k AnotherKey) Equals(other Hashable) bool {
	o, ok := other.(AnotherKey)
	return ok && k == o
}

func TestHashMapBasic(t *testing.T) {
	t.Run("InsertAndGet", func(t *testing.T) {
		hm := NewHashMap[string](WithCapacity(8))
		key := TestKey{1, "a"}
		hm.Set(key, "value1")

		// 测试正常获取
		val, exists := hm.Get(key)
		assert.True(t, exists)
		assert.Equal(t, "value1", val)

		// 测试不存在key
		_, exists = hm.Get(TestKey{2, "b"})
		assert.False(t, exists)
	})

	t.Run("UpdateValue", func(t *testing.T) {
		hm := NewHashMap[string](WithCapacity(8))
		key := TestKey{1, "a"}
		hm.Set(key, "value1")
		hm.Set(key, "value2")

		val, exists := hm.Get(key)
		assert.True(t, exists)
		assert.Equal(t, "value2", val)
	})

	t.Run("DeleteKey", func(t *testing.T) {
		hm := NewHashMap[string](WithCapacity(8))
		key := TestKey{1, "a"}
		hm.Set(key, "value1")

		// 删除存在的key
		hm.Delete(key)
		assert.Equal(t, 0, hm.Size())

		// 删除不存在的key
		hm.Delete(TestKey{2, "b"})
	})
}

func TestHashCollision(t *testing.T) {
	hm := NewHashMap[string](WithCapacity(16))

	// 构造哈希冲突的key
	key1 := TestKey{1, "a"}  // Hash: 1+1=2
	key2 := TestKey{0, "bb"} // Hash: 0+2=2
	key3 := TestKey{2, "a"}  // Hash: 2+1=3

	hm.Set(key1, "value1")
	hm.Set(key2, "value2")
	hm.Set(key3, "value3")

	assert.Equal(t, 3, hm.Size())

	t.Run("GetCollisionKeys", func(t *testing.T) {
		val, exists := hm.Get(key1)
		assert.True(t, exists)
		assert.Equal(t, "value1", val)

		val, exists = hm.Get(key2)
		assert.True(t, exists)
		assert.Equal(t, "value2", val)
	})

	t.Run("DeleteCollisionKey", func(t *testing.T) {
		hm.Delete(key1)
		assert.Equal(t, 2, hm.Size())
		_, exists := hm.Get(key1)
		assert.False(t, exists)
	})
}

func TestAutoResize(t *testing.T) {
	initialCap := 16
	hm := NewHashMap[int](WithCapacity(initialCap))

	// 插入足够数据触发扩容 (16 * 0.75 = 12)
	for i := 0; i < 13; i++ {
		key := TestKey{i, ""}
		hm.Set(key, i)
	}

	// 验证扩容后的容量
	assert.Greater(t, len(hm.buckets), initialCap)

	// 验证所有数据仍然可访问
	for i := 0; i < 13; i++ {
		val, exists := hm.Get(TestKey{i, ""})
		assert.True(t, exists)
		assert.Equal(t, i, val)
	}
}

func TestConcurrency(t *testing.T) {
	hm := NewHashMap[int](WithCapacity(32))
	var wg sync.WaitGroup

	// 并发写入测试
	numWorkers := 100
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(n int) {
			defer wg.Done()
			key := TestKey{n, "test"}
			hm.Set(key, n)
			hm.Get(key)
		}(i)
	}

	// 并发删除测试
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(n int) {
			defer wg.Done()
			key := TestKey{n, "test"}
			hm.Delete(key)
		}(i)
	}

	wg.Wait()
}

func TestTypeSafety(t *testing.T) {
	hm := NewHashMap[string](WithCapacity(8))

	// 不同类型但哈希值相同
	key1 := TestKey{1, "a"} // Hash = 2
	key2 := AnotherKey(2)   // Hash = 2

	hm.Set(key1, "value1")
	hm.Set(key2, "value2")

	// 验证不同类型不会冲突
	val, exists := hm.Get(key1)
	assert.True(t, exists)
	assert.Equal(t, "value1", val)

	val, exists = hm.Get(key2)
	assert.True(t, exists)
	assert.Equal(t, "value2", val)
}

func TestEdgeCases(t *testing.T) {
	t.Run("NilKey", func(t *testing.T) {
		hm := NewHashMap[string](WithCapacity(8))
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with nil key")
			}
		}()

		// 测试nil key
		hm.Set(nil, "value")
	})

	t.Run("ZeroCapacity", func(t *testing.T) {
		hm := NewHashMap[string](WithCapacity(0))
		assert.Equal(t, 1, len(hm.buckets))
	})

	t.Run("DuplicateInsert", func(t *testing.T) {
		hm := NewHashMap[string](WithCapacity(8))
		key := TestKey{1, "a"}
		hm.Set(key, "v1")
		hm.Set(key, "v2")
		assert.Equal(t, 1, hm.Size())
	})
}
