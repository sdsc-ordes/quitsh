package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	stack := Stack[string]{}

	stack.Push("a")
	assert.Equal(t, 1, stack.Len())

	stack.Push("b")
	assert.Equal(t, 2, stack.Len())

	assert.Equal(t, "b", stack.Pop())
	assert.Equal(t, 1, stack.Len())

	stack.Push("c")
	assert.Equal(t, 2, stack.Len())

	assert.Equal(t, "c", stack.Pop())
	assert.Equal(t, 1, stack.Len())

	assert.Equal(t, "a", stack.Pop())
	assert.Equal(t, 0, stack.Len())

	assert.Panics(t, func() { _ = stack.Pop() })
}

func TestStackFrontPop(t *testing.T) {
	stack := NewStackWithCap[string](10)

	stack.Push("a")
	assert.Equal(t, 1, stack.Len())

	stack.Push("b")
	assert.Equal(t, 2, stack.Len())

	assert.Equal(t, "a", stack.PopFront())
	assert.Equal(t, 1, stack.Len())

	stack.Push("c")
	assert.Equal(t, 2, stack.Len())

	assert.Equal(t, "b", stack.PopFront())
	assert.Equal(t, 1, stack.Len())

	assert.Equal(t, "c", stack.Pop())
	assert.Equal(t, 0, stack.Len())

	assert.Panics(t, func() { _ = stack.PopFront() })
}

func TestStackTraverse(t *testing.T) {
	stack := NewStackWithCap[string](10)
	stack.Push("a", "b", "c")

	seq := []string{}

	stack.Visit(func(_ int, s string) bool {
		seq = append(seq, s)

		return true
	})

	assert.Equal(t, []string{"c", "b", "a"}, seq)

	seq = []string{}

	stack.Visit(func(_ int, s string) bool {
		seq = append(seq, s)

		return s != "b"
	})
	assert.Equal(t, []string{"c", "b"}, seq)
}

func TestStackTraverseUpward(t *testing.T) {
	stack := NewStackWithCap[string](10)
	stack.Push("a", "b", "c")

	seq := []string{}

	stack.VisitUpward(func(_ int, s string) bool {
		seq = append(seq, s)

		return true
	})

	assert.Equal(t, []string{"a", "b", "c"}, seq)

	seq = []string{}

	stack.VisitUpward(func(_ int, s string) bool {
		seq = append(seq, s)

		return s != "b"
	})
	assert.Equal(t, []string{"a", "b"}, seq)
}
