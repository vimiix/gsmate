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

type node[T comparable] struct {
	value T
	next  *node[T]
}

type LinkedList[T comparable] struct {
	head *node[T]
	tail *node[T]
	size int
}

func NewLinkedList[T comparable]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// Len returns the number of elements in the linked list.
func (l *LinkedList[T]) Len() int {
	return l.size
}

// Push adds a new element to the end of the linked list.
func (l *LinkedList[T]) Push(v T) {
	n := &node[T]{value: v}
	if l.head == nil {
		l.head = n
		l.tail = n
	} else {
		l.tail.next = n
		l.tail = n
	}
	l.size++
}

// Pop removes and returns the value at the front of the linked list.
func (l *LinkedList[T]) Pop() any {
	if l.head == nil {
		return nil
	}
	v := l.head.value
	l.head = l.head.next
	l.size--
	return v
}

// Values returns all values in the linked list.
func (l *LinkedList[T]) Values() []T {
	rs := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.next {
		rs = append(rs, n.value)
	}
	return rs
}

// IsEmpty returns true if the linked list is empty.
func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}

// Clear removes all elements from the linked list.
func (l *LinkedList[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

// Remove removes the first occurrence of a specific value from the linked list.
func (l *LinkedList[T]) Remove(v T) {
	if l.head == nil {
		return
	}
	if l.head.value == v {
		l.head = l.head.next
		l.size--
		return
	}
	n := l.head
	for n.next != nil {
		if n.next.value == v {
			n.next = n.next.next
			l.size--
			return
		}
		n = n.next
	}
}

// Range calls the function f for each element in the linked list.
func (l *LinkedList[T]) Range(f func(T)) {
	for n := l.head; n != nil; n = n.next {
		f(n.value)
	}
}
