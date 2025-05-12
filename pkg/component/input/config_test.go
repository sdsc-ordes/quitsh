package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelativeDir(t *testing.T) {
	t.Parallel()
	type Test struct {
		Path     string
		Expected string
		Trimmed  bool
	}

	tests := []Test{
		{Path: "/repo/component/a/b", Expected: "a/b", Trimmed: true},
		{Path: "/repo/component/a//b/", Expected: "a/b", Trimmed: true},
		{Path: "/repo/aa/a//b/", Expected: "/repo/aa/a/b", Trimmed: false},
		{Path: "/repo/component-a", Expected: "/repo/component-a", Trimmed: false},
	}

	for _, test := range tests {
		p, trimmed := BaseDir("/repo/component").TrimOffFrom(test.Path)
		assert.Equal(t, test.Expected, p)
		assert.Equal(t, test.Trimmed, trimmed)
	}
}
