package syncmap

import (
	"sync"
)

type Map[K, V any] struct {
	state sync.Map
}

func New[K, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

// Set stores the value for a key.
func (m *Map[K, V]) Set(key K, value V) {
	m.state.Store(key, value)
}

func (m *Map[K, V]) zeroVal() V {
	var tmp V
	return tmp
}

// Get returns the value stored in the map for a key, or zero value if key is not present.
// The ok result indicates whether value was found in the map.
func (m *Map[K, V]) Get(key K) (V, bool) {
	v, ok := m.state.Load(key)
	if !ok {
		return m.zeroVal(), false
	}
	return v.(V), true
}

func (m *Map[K, V]) Contains(key K) bool {
	_, ok := m.state.Load(key)
	return ok
}

// Delete deletes the value for a key.
func (m *Map[K, V]) Delete(key K) {
	m.state.Delete(key)
}

// GetAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[K, V]) GetAndDelete(key K) (actual V, loaded bool) {
	actual, loaded = m.Get(key)
	m.state.Delete(key)
	return
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is the nil interface value).
func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.state.CompareAndDelete(key, old)
}

// Clear deletes all the entries, resulting in an empty Map.
func (m *Map[K, V]) Clear() {
	m.state = sync.Map{}
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[K, V]) GetOrSet(key K, value V) (actual V, loaded bool) {
	v, loaded := m.state.LoadOrStore(key, value)
	return v.(V), loaded
}

// Swap swaps the value for a key and returns the previous value if any.
// The swapped result reports whether the key was present.
func (m *Map[K, V]) Swap(key K, val V) (existing V, swapped bool) {
	old, replaced := m.state.Swap(key, val)
	if replaced {
		return old.(V), true
	}
	return m.zeroVal(), false
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
func (m *Map[K, V]) CompareAndSwap(key K, old, new V) (swapped bool) {
	return m.state.CompareAndSwap(key, old, new)
}

// Keys returns a slice of all keys present in the map.
func (m *Map[K, V]) Keys() []K {
	keys := make([]K, 0, m.Len())
	m.state.Range(func(key, _ any) bool {
		keys = append(keys, key.(K))
		return true
	})
	return keys
}

// Values returns a slice of all values present in the map.
func (m *Map[K, V]) Values() []V {
	values := make([]V, 0, m.Len())
	m.state.Range(func(_ interface{}, value any) bool {
		values = append(values, value.(V))
		return true
	})
	return values
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.state.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (m *Map[K, V]) Len() int {
	count := 0
	m.state.Range(func(_ interface{}, _ interface{}) bool {
		count++
		return true
	})
	return count
}
