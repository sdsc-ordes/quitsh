package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyMerge(t *testing.T) {
	src := map[int]int{
		1: 99,
		2: 98,
		3: 4,
	}

	dest := map[int]int{
		1: 100,
		2: 101,
	}

	existing := Merge(dest, src)
	assert.Contains(t, existing, 1)
	assert.Contains(t, existing, 2)
	assert.Equal(t, 99, dest[1])
	assert.Equal(t, 98, dest[2])
}
