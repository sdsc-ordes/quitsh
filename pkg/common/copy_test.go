package common

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

// CopySlice copies a slice.
func TestCopySlice(t *testing.T) {
	c := []string{"a", "b"}
	r := CopySlice(c)
	c[1] = "c"

	assert.Equal(t, "b", r[1])

	c = []string{"a", "b"}
	r = CopySliceC(c, 4)
	c[1] = "c"

	assert.Equal(t, "b", r[1])
}
