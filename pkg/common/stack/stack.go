package stack

import (
	"slices"

	"github.com/sdsc-ordes/quitsh/pkg/debug"
)

type Stack[T any] struct {
	stack []T
}

// NewStack creates a new stack with passing `capacity` argument
// to `make(T[],...)`.
func NewStack[T any]() Stack[T] {
	return Stack[T]{}
}

// NewStack creates a new stack with passing `capacity` argument
// to `make(T[],...)`.
func NewStackWithCap[T any](capacity int) Stack[T] {
	return Stack[T]{stack: make([]T, 0, capacity)}
}

// Pop pops the top element on the stack.
// The stack size needs to be greater > 0.
func (s *Stack[T]) Pop() T {
	res := s.Top()
	s.stack = s.stack[:len(s.stack)-1]

	return res
}

// Returns the top element on the stack.
func (s *Stack[T]) Top() T {
	debug.Assert(s.Len() != 0, "the stack size is not > 0")

	return s.stack[len(s.stack)-1]
}

// PopFront pops the bottom level on the stack.
// This method is useful to do Breath-First-Traversal.
// instead of Depth-First-Traversal when using `Pop`.
func (s *Stack[T]) PopBottom() T {
	res := s.Bottom()
	s.stack = s.stack[1:]

	return res
}

// Returns the bottom element.
func (s *Stack[T]) Bottom() T {
	debug.Assert(s.Len() != 0, "the stack size is not > 0")

	return s.stack[0]
}

// Visit travers the stack from top to bottom and applies a function.
// If the visitor returns `false` the iteration is aborted.
func (s *Stack[T]) Visit(visitor func(int, T) bool) {
	for i, el := range slices.Backward(s.stack) {
		if !visitor(i, el) {
			return
		}
	}
}

// Visit travers the stack from bottom to top and applies a function.
// If the visitor returns `false` the iteration is aborted.
func (s *Stack[T]) VisitUpward(visitor func(int, T) bool) {
	for i, el := range s.stack {
		if !visitor(i, el) {
			return
		}
	}
}

// Push appends an element to the stack.
func (s *Stack[T]) Push(t ...T) {
	s.stack = append(s.stack, t...)
}

// Len returns the length of the stack.
func (s *Stack[T]) Len() int {
	return len(s.stack)
}
