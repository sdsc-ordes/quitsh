package set

import (
	"fmt"
	"iter"
	"maps"
)

type Unordered[T comparable] struct {
	set map[T]struct{}
}

// NewUnordered constructs an `Unordered` set from elements.
// The capacity is set to `2*len(el)`.
func NewUnordered[T comparable](el ...T) (s Unordered[T]) {
	if len(el) == 0 {
		s.set = make(map[T]struct{})
	} else {
		s.set = make(map[T]struct{}, 2*len(el)) //nolint:mnd
	}

	for i := range el {
		s.Insert(el[i])
	}

	return s
}

// NewUnorderedWithCap constructs an `Unordered` set from elements
// and sets the capacity.
func NewUnorderedWithCap[T comparable](capacity int, el ...T) (s Unordered[T]) {
	capacity = max(capacity, len(el))
	s.set = make(map[T]struct{}, capacity)

	for i := range el {
		s.Insert(el[i])
	}

	return s
}

// Values returns an iterator over values in m.
// The iteration order is not specified and is not guaranteed
// to be the same from one call to the next.
func (s *Unordered[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s.set {
			if !yield(k) {
				return
			}
		}
	}
}

// Keys returns the same as `Values`.
func (s *Unordered[T]) Keys() iter.Seq[T] {
	return s.Values()
}

// Collect collects values from seq into a new `Unordered` set.
func Collect[E comparable](seq iter.Seq[E]) (res Unordered[E]) {
	res = NewUnordered[E]()

	for s := range seq {
		res.Insert(s)
	}

	return res
}

// Equal tests if two sets are identical.
func (s *Unordered[T]) Equal(other *Unordered[T]) bool {
	return maps.Equal(s.set, other.set)
}

// Insert inserts an element into the set and returns `true` if
// it existed already.
func (s *Unordered[T]) Insert(e T) bool {
	if s.set == nil {
		s.set = make(map[T]struct{})
	}

	if _, exists := s.set[e]; exists {
		return true
	}

	s.set[e] = struct{}{}

	return false
}

// Removes an element `e` from the set and returns
// `true` if it was removed.
func (s *Unordered[T]) Remove(e T) bool {
	if s.set == nil {
		s.set = make(map[T]struct{})
	}

	if _, exists := s.set[e]; !exists {
		return false
	}

	delete(s.set, e)

	return true
}

// Exists checks if `e` is already in the set.
func (s *Unordered[T]) Exists(e T) bool {
	if s.set == nil {
		s.set = make(map[T]struct{})
	}

	_, exists := s.set[e]

	return exists
}

// Len returns the length of the set.
func (s *Unordered[T]) Len() int {
	return len(s.set)
}

// String returns the conversion to a string.
func (s *Unordered[T]) String() string {
	return fmt.Sprintf("%v", s.set)
}
