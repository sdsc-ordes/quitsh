package nix

import (
	"github.com/sdsc-ordes/quitsh/pkg/exec"
)

type EvalOption func(cmd *[]string) error

// WithEvalImpure uses `impure` evaluation on `Eval`.
func WithEvalImpure() EvalOption {
	return func(cmd *[]string) error {
		(*cmd) = append((*cmd), "--impure")

		return nil
	}
}

// WithEvalOutputRaw uses `raw` output on `Eval`.
func WithEvalOutputRaw() EvalOption {
	return func(cmd *[]string) error {
		(*cmd) = append((*cmd), "--raw")

		return nil
	}
}

// WithEvalOutputRaw uses JSON output on `Eval`.
func WithEvalOutputJSON() EvalOption {
	return func(cmd *[]string) error {
		(*cmd) = append((*cmd), "--json")

		return nil
	}
}

// EvalTemplate evaluates a Nix expression and returns the result.
func EvalTemplate(
	nixx *exec.CmdContext,
	temp string,
	data any,
	opts ...EvalOption) (string, error) {
	run := func(c *exec.CmdContext, file string) (string, error) {
		cmd := []string{"eval", "--file", file}

		for i := range opts {
			e := opts[i](&cmd)
			if e != nil {
				return "", e
			}
		}

		return c.Get(cmd...)
	}

	return exec.WithTemplate(nixx, temp, data, run)
}
