package nix

import (
	"github.com/sdsc-ordes/quitsh/pkg/exec"
)

type EvalOption func(*opts) error

type opts struct {
	addArgs []string
}

func (c *opts) Apply(options ...EvalOption) error {
	for _, f := range options {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

// WithEvalImpure uses `impure` evaluation on `Eval`.
func WithEvalImpure() EvalOption {
	return func(o *opts) error {
		o.addArgs = append(o.addArgs, "--impure")

		return nil
	}
}

// WithEvalOutputRaw uses `raw` output on `Eval`.
func WithEvalOutputRaw() EvalOption {
	return func(o *opts) error {
		o.addArgs = append(o.addArgs, "--raw")

		return nil
	}
}

// WithEvalOutputJSON uses JSON output on `Eval`.
func WithEvalOutputJSON() EvalOption {
	return func(o *opts) error {
		o.addArgs = append(o.addArgs, "--json")

		return nil
	}
}

// WithEvalNoCache uses no eval cache.
// Note: Good when used in parallel to reduce SQLite contention.
func WithEvalNoCache() EvalOption {
	return func(o *opts) error {
		o.addArgs = append(o.addArgs, "--no-eval-cache")

		return nil
	}
}

// EvalTemplate evaluates a Nix expression and returns the result.
func EvalTemplate(
	nixx *exec.CmdContext,
	temp string,
	data any,
	option ...EvalOption) (string, error) {
	var o opts
	o.Apply(option...)

	run := func(c *exec.CmdContext, file string) (string, error) {
		cmd := []string{"eval", "--file", file}
		cmd = append(cmd, o.addArgs...)

		return c.Get(cmd...)
	}

	return exec.WithTemplate(nixx, temp, data, run)
}
