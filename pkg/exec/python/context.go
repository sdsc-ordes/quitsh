package python

import (
	"github.com/sdsc-ordes/quitsh/pkg/exec"
)

// NewVEnvCtxBuilder creates a new command ctx buildter for a python virtual environment.
// It will enable path look up for passed commands to find the executable in the
// modified env. `PATH` and adds `VIRTUAL_ENV` env. variable.
// NewCtx returns a new Go command context builder.
func NewVEnvCtxBuilder(venvDir string, env []string) exec.CmdContextBuilder {
	return exec.NewCmdCtxBuilder().
		Env("VIRTUAL_ENV=" + venvDir).
		Env(env...).
		Paths(venvDir)
}
