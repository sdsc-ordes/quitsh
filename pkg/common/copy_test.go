//go:build test && (test_small || test_all)

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// CopySlice copies a slice.
func TestCopySlice(t *testing.T) {
	t.Parallel()

	var c []string
	r := CopySlice(c)
	assert.Nil(t, r)

	r = CopySliceC(c, 10)
	assert.Nil(t, r)

	c = []string{"a", "b"}
	r = CopySlice(c)
	assert.Len(t, r, 2)
	c[1] = "c"
	assert.Equal(t, "b", r[1])

	c = []string{"a", "b"}
	r = CopySliceC(c, 4)
	assert.Len(t, r, 2)
	c[1] = "c"
	assert.Equal(t, "b", r[1])
}
