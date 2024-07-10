// Copyright 2024 Qian Yao
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orderedmap

import "gsmate/internal/linkedlist"

type OrderedMap[K comparable, V any] struct {
	keys *linkedlist.LinkedList[K]
	m    map[K]V
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		m:    make(map[K]V),
		keys: linkedlist.NewLinkedList[K](),
	}
}

// Keys returns the keys of the map in the order of the values.
func (m *OrderedMap[K, V]) Keys() []K {
	return m.keys.Values()
}

// Values returns the values of the map in the order of the keys.
func (m *OrderedMap[K, V]) Values() []V {
	rs := make([]V, 0, m.keys.Len())
	m.keys.Range(func(k K) {
		rs = append(rs, m.m[k])
	})
	return rs
}

// Len returns the number of elements in the map.
func (m *OrderedMap[K, V]) Len() int {
	return m.keys.Len()
}

// Set sets the value associated with the given key.
func (m *OrderedMap[K, V]) Set(k K, v V) {
	if _, ok := m.m[k]; !ok {
		m.keys.Push(k)
	}
	m.m[k] = v
}

// Get returns the value associated with the given key.
func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.m[k]
	return v, ok
}

// Delete removes the value associated with the given key.
func (m *OrderedMap[K, V]) Delete(k K) {
	m.keys.Remove(k)
	delete(m.m, k)
}

// Range calls f for each element in the map in the order of the keys.
func (m *OrderedMap[K, V]) Range(f func(k K, v V)) {
	m.keys.Range(func(k K) {
		f(k, m.m[k])
	})
}

// Clear removes all elements from the map.
func (m *OrderedMap[K, V]) Clear() {
	m.keys.Clear()
	m.m = make(map[K]V)
}

// IsEmpty returns true if the map is empty.
func (m *OrderedMap[K, V]) IsEmpty() bool {
	return m.Len() == 0
}
