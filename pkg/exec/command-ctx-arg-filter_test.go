package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialArgsFilter(t *testing.T) {
	type Data struct {
		Input  []string
		Expect []string
	}

	tests := []Data{}
	for _, arg := range defaultCredentialArgs {
		tests = append(tests, []Data{
			{
				Input:  []string{"-a", "--" + arg, "--banana", "password"},
				Expect: []string{"-a", "--" + arg, hiddenDefault, "password"},
			},
			{
				Input:  []string{"-a", "--db-" + arg, "banana", "password"},
				Expect: []string{"-a", "--db-" + arg, hiddenDefault, "password"},
			},
			{
				Input:  []string{"password", "--db-" + arg + "-write=banana", "password"},
				Expect: []string{"password", "--db-" + arg + "-write=" + hiddenDefault, "password"},
			},
			{
				Input:  []string{"-a", "-" + arg, "--banana", "password"},
				Expect: []string{"-a", "-" + arg, hiddenDefault, "password"},
			},
			{
				Input:  []string{"-a", "-db-" + arg, "banana", "password"},
				Expect: []string{"-a", "-db-" + arg, hiddenDefault, "password"},
			},
			{
				Input:  []string{"password", "-db-" + arg + "-write=banana", "password"},
				Expect: []string{"password", "-db-" + arg + "-write=" + hiddenDefault, "password"},
			},
			{
				Input:  []string{"-a", "-" + arg + "========", "--"},
				Expect: []string{"-a", "-" + arg + "=" + hiddenDefault, "--"},
			}}...)
	}

	for _, test := range tests {
		res := DefaultCredentialFilter(test.Input)
		assert.Equal(t, test.Expect, res)
	}
}
