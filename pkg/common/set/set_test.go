package set

import (
	"testing"

	"deedles.dev/xiter"
	"github.com/stretchr/testify/assert"
)

func TestUnorderedSetDefault(t *testing.T) {
	t.Parallel()
	s := Unordered[int]{}
	assert.Nil(t, s.set)
	assert.False(t, s.Exists(10))
	assert.NotNil(t, s.set)
	assert.Empty(t, s.set)

	s = Unordered[int]{}
	assert.False(t, s.Insert(10))
	assert.NotNil(t, s.set)
	assert.Len(t, s.set, 1)

	s = Unordered[int]{}
	assert.False(t, s.Remove(10))
	assert.NotNil(t, s.set)
	assert.Empty(t, s.set)
}

func TestUnorderedSetNew(t *testing.T) {
	t.Parallel()
	s := NewUnorderedWithCap[int](10)
	assert.False(t, s.Exists(10))
	assert.Empty(t, s.set)
	assert.Equal(t, 0, s.Len())

	s = NewUnordered(1, 2, 3, 4)
	assert.True(t, s.Exists(1))
	assert.True(t, s.Exists(2))
	assert.True(t, s.Exists(3))
	assert.True(t, s.Exists(4))
	assert.Equal(t, 4, s.Len())
}

func TestUnorderedSetInsert(t *testing.T) {
	t.Parallel()
	s := NewUnordered(1, 2, 3, 4)
	assert.True(t, s.Exists(1))
	assert.True(t, s.Exists(2))
	assert.True(t, s.Exists(3))
	assert.True(t, s.Exists(4))

	exists := s.Insert(1)
	assert.True(t, exists)
	assert.Len(t, s.set, 4)

	exists = s.Insert(5)
	assert.False(t, exists)
	assert.Len(t, s.set, 5)

	removed := s.Remove(10)
	assert.False(t, removed)
	assert.Len(t, s.set, 5)

	removed = s.Remove(5)
	assert.True(t, removed)
	assert.Len(t, s.set, 4)
}

func TestIteration(t *testing.T) {
	t.Parallel()
	expect := NewUnordered(1, 4, 9, 16)
	s := NewUnordered(1, 2, 3, 4)

	res := Collect(xiter.Map(s.Values(), func(v int) int { return v * v }))

	assert.True(t, expect.Equal(&res))
	assert.Equal(t, expect, res)
}
