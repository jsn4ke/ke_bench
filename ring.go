// Package kebench provides a Ring data structure implementation.
package kebench

import (
	"sync"
)

// Ring is a circular buffer data structure that can hold elements of any type.
type Ring[T any] struct {
	size int

	head, tail int

	container []T

	mtx sync.Mutex
}

// NewRing creates a new Ring with the specified size.
func NewRing[T any](size int) *Ring[T] {
	return &Ring[T]{
		size:      size,
		container: make([]T, size+1), // Waste one space for easy detection of full and empty states
	}
}

// full checks if the Ring is full.
func (r *Ring[T]) full() bool {
	return r.tail == r.inc(r.head)
}

// empty checks if the Ring is empty.
func (r *Ring[T]) empty() bool {
	return r.tail == r.head
}

// inc increments the given index in a circular manner.
func (r *Ring[T]) inc(in int) int {
	return (in + 1) % int(r.size+1)
}

// Push adds an element to the Ring. Returns true if successful, false if the Ring is full.
func (r *Ring[T]) Push(t T) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.full() {
		return false
	}
	r.container[r.head] = t
	r.head = r.inc(r.head)
	return true
}

// Get retrieves an element from the Ring. Returns the element and true if successful, or a zero value and false if the Ring is empty.
func (r *Ring[T]) Get() (t T, ok bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.empty() {
		return
	}
	t = r.container[r.tail]
	r.tail = r.inc(r.tail)
	return t, true
}
