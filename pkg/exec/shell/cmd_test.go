package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdString(t *testing.T) {
	t.Parallel()
	type D struct {
		expect string
		input  []string
	}

	tests := []D{
		{expect: "'a' 'b' 'c'", input: []string{"a", "b", "c"}},
		{expect: `'a'\''a' 'b' 'c'`, input: []string{"a'a", "b", "c"}},
		{
			expect: `'a'\'''\'''\''a' ''\'''\''b'\'''\''' 'c'`,
			input:  []string{"a'''a", "''b''", "c"},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expect, CmdToString(test.input...))
	}
}

func TestCmdStringF(t *testing.T) {
	t.Parallel()
	type D struct {
		expect string
		input  []string
	}

	tests := []D{
		{expect: "banana 'a' 'b' 'c'", input: []string{"a", "b", "c"}},
		{expect: `banana 'a'\''a' 'b' 'c'`, input: []string{"a'a", "b", "c"}},
		{
			expect: `banana 'a'\'''\'''\''a' ''\'''\''b'\'''\''' 'c'`,
			input:  []string{"a'''a", "''b''", "c"},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expect, CmdToStringF("banana", test.input...))
	}
}
