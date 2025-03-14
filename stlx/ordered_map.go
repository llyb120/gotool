package stlx

import (
	"sync"
)

// OrderedMap 是一个协程安全的有序映射，按插入顺序维护键值对
type OrderedMap[K comparable, V any] struct {
	mu      sync.RWMutex
	keys    []K
	values  []V
	indexes map[K]int
}

// NewOrderedMap 创建一个新的有序映射
func NewMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		indexes: make(map[K]int),
	}
}

// Set 添加或更新键值对
func (om *OrderedMap[K, V]) Set(key K, value V) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.set(key, value)
}

// Get 获取键对应的值
func (om *OrderedMap[K, V]) Get(key K) (V, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if index, exists := om.indexes[key]; exists {
		return om.values[index], true
	}

	var zero V
	return zero, false
}

// Del 删除键值对
func (om *OrderedMap[K, V]) Del(key K) V {
	om.mu.Lock()
	defer om.mu.Unlock()

	pos, exists := om.indexes[key]
	if !exists {
		var zero V
		return zero
	} else {
		delete(om.indexes, key)
		val := om.values[pos]
		om.keys = append(om.keys[:pos], om.keys[pos+1:]...)
		om.values = append(om.values[:pos], om.values[pos+1:]...)
		return val
	}
}

// Size 返回映射大小
func (om *OrderedMap[K, V]) Len() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.keys)
}

// Keys 按插入顺序返回所有键
func (om *OrderedMap[K, V]) Keys() []K {
	om.mu.RLock()
	defer om.mu.RUnlock()

	keys := make([]K, len(om.keys))
	copy(keys, om.keys)
	return keys
}

// Vals 按插入顺序返回所有值
func (om *OrderedMap[K, V]) Vals() []V {
	om.mu.RLock()
	defer om.mu.RUnlock()

	values := make([]V, len(om.values))
	copy(values, om.values)
	return values
}

// Clear 清空映射
func (om *OrderedMap[K, V]) Clear() {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.clear()
}

// For 按顺序遍历所有键值对
func (om *OrderedMap[K, V]) For(fn func(key K, value V) bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	for i, key := range om.keys {
		if !fn(key, om.values[i]) {
			break
		}
	}
}
