package tags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestMatchExpr_Matches(t *testing.T) {
	tests := []struct {
		name     string
		expr     string // the raw build tag expression
		tags     []Tag  // passed to Matches
		expected bool   // expected result
	}{
		{
			name:     "simple match",
			expr:     "",
			tags:     []Tag{},
			expected: true,
		},
		{
			name:     "simple match",
			expr:     "",
			tags:     []Tag{NewTag("a"), NewTag("b"), NewTag("c")},
			expected: true,
		},
		{
			name:     "simple match",
			expr:     "a",
			tags:     []Tag{NewTag("a")},
			expected: true,
		},
		{
			name:     "simple no match",
			expr:     "a",
			tags:     []Tag{NewTag("c")},
			expected: false,
		},
		{
			name:     "negated match",
			expr:     "!c",
			tags:     []Tag{NewTag("a")},
			expected: true,
		},
		{
			name:     "negated no match",
			expr:     "!c",
			tags:     []Tag{NewTag("c")},
			expected: false,
		},
		{
			name:     "compound match",
			expr:     "a && b",
			tags:     []Tag{NewTag("a"), NewTag("b")},
			expected: true,
		},
		{
			name:     "compound no match",
			expr:     "a && b",
			tags:     []Tag{NewTag("a")},
			expected: false,
		},
		{
			name:     "or match",
			expr:     "a || c",
			tags:     []Tag{NewTag("c")},
			expected: true,
		},
		{
			name:     "complex expression match",
			expr:     "(a && !b) || c",
			tags:     []Tag{NewTag("c")},
			expected: true,
		},
		{
			name:     "complex expression match",
			expr:     "(a && !b) || c",
			tags:     []Tag{NewTag("a"), NewTag("b")},
			expected: false,
		},
		{
			name:     "complex expression match",
			expr:     "(a && !b) || c",
			tags:     []Tag{NewTag("a")},
			expected: true,
		},
		{
			name:     "complex expression no match",
			expr:     "(a && b) || c",
			tags:     []Tag{NewTag("d")},
			expected: false,
		},
		{
			name:     "complex expression match with '-'",
			expr:     "(do-this && !do.that) || al-wa-ys",
			tags:     []Tag{NewTag("do-this")},
			expected: true,
		},
		{
			name:     "complex expression match with '-'",
			expr:     "(do-this && !do.that) || al-wa-ys",
			tags:     []Tag{NewTag("al.wa.ys")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			me, err := NewExpr(tt.expr)
			require.NoError(t, err)

			result := me.Matches(tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}
