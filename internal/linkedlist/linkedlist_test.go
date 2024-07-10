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

package linkedlist

import "testing"

func TestPush(t *testing.T) {
	// Test pushing a single element
	l := NewLinkedList[int]()
	l.Push(1)
	if l.Len() != 1 {
		t.Errorf("Expected length to be 1, got %d", l.Len())
	}
	if l.head.value != 1 {
		t.Errorf("Expected head value to be 1, got %d", l.head.value)
	}
	if l.tail.value != 1 {
		t.Errorf("Expected tail value to be 1, got %d", l.tail.value)
	}

	// Test pushing multiple elements
	l.Push(2)
	if l.Len() != 2 {
		t.Errorf("Expected length to be 2, got %d", l.Len())
	}
	if l.tail.value != 2 {
		t.Errorf("Expected tail value to be 2, got %d", l.tail.value)
	}
	if l.tail.next != nil {
		t.Error("Expected tail.next to be nil")
	}
}

func TestLinkedListPop(t *testing.T) {
	l := NewLinkedList[int]()
	if l.Pop() != nil {
		t.Error("Expected nil, got value")
	}

	l.Push(1)
	r := l.Pop()
	if r != 1 {
		t.Errorf("Expected 1, got value %v", r)
	}

	l.Push(2)
	r = l.Pop()
	if r != 2 {
		t.Errorf("Expected 2, got value %v", r)
	}

	if l.Pop() != nil {
		t.Error("Expected nil, got value")
	}

	l.Push(1)
	l.Push(2)
	l.Push(3)
	r = l.Pop()
	if r != 1 {
		t.Errorf("Expected 1, got value %v", r)
	}
	r = l.Pop()
	if r != 2 {
		t.Errorf("Expected 2, got value %v", r)
	}
	r = l.Pop()
	if r != 3 {
		t.Errorf("Expected 3, got value %v", r)
	}
	r = l.Pop()
	if r != nil {
		t.Errorf("Expected nil, got value %v", r)
	}
}

func TestRemove(t *testing.T) {
	// Test removing the only element
	l := NewLinkedList[int]()
	l.Push(1)
	l.Remove(1)
	if !l.IsEmpty() {
		t.Error("Expected linked list to be empty after removing the only element")
	}

	// Test removing the first element
	l.Push(1)
	l.Push(2)
	l.Remove(1)
	if l.Len() != 1 {
		t.Errorf("Expected length to be 1, got %d", l.Len())
	}
	if l.head.value != 2 {
		t.Errorf("Expected head value to be 2, got %d", l.head.value)
	}

	// Test removing a middle element
	l.Push(3)
	l.Remove(2)
	if l.Len() != 1 {
		t.Errorf("Expected length to be 1, got %d", l.Len())
	}
	if l.head.value != 3 {
		t.Errorf("Expected head value to be 3, got %d", l.head.value)
	}

	// Test removing the last element
	l.Push(4)
	l.Remove(3)
	if l.Len() != 1 {
		t.Errorf("Expected length to be 1, got %d", l.Len())
	}
	if l.head.value != 4 {
		t.Errorf("Expected head value to be 4, got %d", l.head.value)
	}
}

func TestLinkedList_Values(t *testing.T) {
	// Test when the linked list is empty
	l := NewLinkedList[int]()
	values := l.Values()
	if len(values) != 0 {
		t.Errorf("Expected empty values slice, got %v", values)
	}

	// Test when the linked list has one element
	l.Push(1)
	values = l.Values()
	if len(values) != 1 || values[0] != 1 {
		t.Errorf("Expected values slice with [1], got %v", values)
	}

	// Test when the linked list has multiple elements
	l.Push(2)
	l.Push(3)
	values = l.Values()
	expected := []int{1, 2, 3}
	if len(values) != len(expected) {
		t.Errorf("Expected values slice with %v, got %v", expected, values)
	} else {
		for i := range expected {
			if values[i] != expected[i] {
				t.Errorf("Expected value at index %d to be %d, got %d", i, expected[i], values[i])
			}
		}
	}
}

func TestClear(t *testing.T) {
	l := NewLinkedList[int]()
	l.Push(1)
	l.Push(2)
	l.Clear()

	if l.head != nil {
		t.Error("Expected head to be nil after Clear")
	}

	if l.tail != nil {
		t.Error("Expected tail to be nil after Clear")
	}

	if l.Len() != 0 {
		t.Errorf("Expected size to be 0 after Clear, got %d", l.Len())
	}
}

func TestRange(t *testing.T) {
	// Test when the linked list is empty
	l := NewLinkedList[int]()
	var count int
	l.Range(func(val int) {
		count++
	})
	if count != 0 {
		t.Errorf("Expected count to be 0 for empty linked list, got %d", count)
	}

	// Test when the linked list has one element
	l.Push(1)
	count = 0
	l.Range(func(val int) {
		count++
		if val != 1 {
			t.Errorf("Expected value to be 1, got %d", val)
		}
	})
	if count != 1 {
		t.Errorf("Expected count to be 1 for single element linked list, got %d", count)
	}

	// Test when the linked list has multiple elements
	l.Push(2)
	l.Push(3)
	var sum int
	l.Range(func(val int) {
		sum += val
	})
	if sum != 6 {
		t.Errorf("Expected sum of values to be 6, got %d", sum)
	}
}
