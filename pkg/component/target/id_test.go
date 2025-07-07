package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	id := DefineID("base", "test")

	n, exists := id.Namespace()
	assert.True(t, exists)
	assert.Equal(t, "base", n)
	assert.Equal(t, "test", id.Name())

	id = ID("test")
	n, exists = id.Namespace()
	assert.False(t, exists)
	assert.Empty(t, n)
	assert.Equal(t, "test", id.Name())

	id = ID("base" + NamespaceSeparator)
	n, exists = id.Namespace()
	assert.True(t, exists)
	assert.Equal(t, "base", n)
	assert.Empty(t, id.Name())

	id = DefineID("base::base", "test::test")
	n, exists = id.Namespace()
	assert.True(t, exists)
	assert.Equal(t, "base-base", n)
	assert.Equal(t, "test-test", id.Name())
}
