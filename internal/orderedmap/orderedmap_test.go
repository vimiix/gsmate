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

import (
	"reflect"
	"testing"
)

func TestKeys(t *testing.T) {
	// Test empty map
	m := NewOrderedMap[int, string]()
	keys := m.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys() = %v, want empty slice", keys)
	}

	// Test map with one element
	m.Set(1, "one")
	keys = m.Keys()
	if len(keys) != 1 || keys[0] != 1 {
		t.Errorf("Keys() = %v, want [1]", keys)
	}

	// Test map with multiple elements
	m.Set(2, "two")
	m.Set(3, "three")
	keys = m.Keys()
	if len(keys) != 3 || keys[0] != 1 || keys[1] != 2 || keys[2] != 3 {
		t.Errorf("Keys() = %v, want [1, 2, 3]", keys)
	}
}

func TestOrderedMap_Values(t *testing.T) {
	// Test empty map
	m := NewOrderedMap[string, int]()
	if values := m.Values(); len(values) != 0 {
		t.Errorf("Expected empty slice, got %v", values)
	}

	// Test map with one element
	m.Set("a", 1)
	if values := m.Values(); !reflect.DeepEqual(values, []int{1}) {
		t.Errorf("Expected [1], got %v", values)
	}

	// Test map with multiple elements
	m.Set("b", 2)
	m.Set("c", 3)
	if values := m.Values(); !reflect.DeepEqual(values, []int{1, 2, 3}) {
		t.Errorf("Expected [1, 2, 3], got %v", values)
	}
}

func TestOrderedMap_Len(t *testing.T) {
	// Test empty map
	m := NewOrderedMap[int, string]()
	if len := m.Len(); len != 0 {
		t.Errorf("Len() = %d, want 0", len)
	}

	// Test map with one element
	m.Set(1, "one")
	if len := m.Len(); len != 1 {
		t.Errorf("Len() = %d, want 1", len)
	}

	// Test map with multiple elements
	m.Set(2, "two")
	m.Set(3, "three")
	if len := m.Len(); len != 3 {
		t.Errorf("Len() = %d, want 3", len)
	}
}

func TestSet(t *testing.T) {
	m := NewOrderedMap[string, string]()

	// Test setting a new key-value pair
	m.Set("key1", "value1")
	if value, _ := m.Get("key1"); value != "value1" {
		t.Errorf("Expected value %s, but got %s", "value1", value)
	}

	// Test updating an existing key with a new value
	m.Set("key1", "updatedValue")
	if value, _ := m.Get("key1"); value != "updatedValue" {
		t.Errorf("Expected value %s, but got %s", "updatedValue", value)
	}

	// Test not pushing the key to keys if it already exists
	initialLen := m.Len()
	m.Set("key1", "anotherValue")
	newLen := m.Len()
	if initialLen != newLen {
		t.Errorf("Expected length %d, but got %d", initialLen, newLen)
	}
}

func TestGet(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("key1", 1)
	m.Set("key2", 2)

	// Test retrieving a value associated with an existing key
	result1, _ := m.Get("key1")
	if result1 != 1 {
		t.Errorf("Expected value for key1 to be 1, got %v", result1)
	}

	// Test retrieving a value associated with a non-existing key
	result2, _ := m.Get("key3")
	if result2 != 0 {
		t.Errorf("Expected value for key3 to be 0, got %v", result2)
	}
}

func TestDelete(t *testing.T) {
	m := NewOrderedMap[int, string]()

	// Test deleting a key that exists in the map
	m.Set(1, "one")
	m.Delete(1)
	if _, ok := m.Get(1); ok {
		t.Error("Expected key 1 to be deleted")
	}

	// Test deleting a key that does not exist in the map
	m.Delete(2)
	if _, ok := m.Get(2); ok {
		t.Error("Expected key 2 to not exist in the map")
	}
}

func TestOrderedMap_Range(t *testing.T) {
	// Test when the map is empty
	m := NewOrderedMap[string, int]()
	var count int
	m.Range(func(k string, v int) {
		count++
	})
	if count != 0 {
		t.Errorf("Expected count to be 0 for an empty map, got %d", count)
	}

	// Test when the map has one element
	m.Set("a", 1)
	count = 0
	m.Range(func(k string, v int) {
		count++
		if k != "a" || v != 1 {
			t.Errorf("Unexpected key-value pair: %s-%d", k, v)
		}
	})
	if count != 1 {
		t.Errorf("Expected count to be 1 for a map with one element, got %d", count)
	}

	// Test when the map has multiple elements
	m.Set("b", 2)
	m.Set("c", 3)
	var sum int
	m.Range(func(k string, v int) {
		sum += v
	})
	if sum != 6 {
		t.Errorf("Expected sum of values to be 6, got %d", sum)
	}
}

func TestClear(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("key1", 1)
	m.Set("key2", 2)

	// Test clearing the map
	m.Clear()

	// Test if the map is empty
	if m.Len() != 0 {
		t.Errorf("Expected length 0, but got %d", m.Len())
	}

	// Test if the keys list is empty
	if m.keys.Len() != 0 {
		t.Errorf("Expected keys list empty, but got %d", m.keys.Len())
	}

	// Test if the map is empty
	if len(m.m) != 0 {
		t.Errorf("Expected map empty, but got %d", len(m.m))
	}
}
