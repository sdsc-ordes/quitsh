package python

import (
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/exec"
)

// NewVEnvCtxBuilder creates a new command ctx builder for a python virtual environment.
// It will enable path look up for passed commands to find the executable in the
// modified env. `PATH` and adds
// `VIRTUAL_ENV` env. variable to point to the virtual env. dir.
// NewCtx returns a new Go command context builder.
func NewVEnvCtxBuilder(venvDir string, env []string) exec.CmdContextBuilder {
	return exec.NewCmdCtxBuilder().
		Env(
			"VIRTUAL_ENV=" + venvDir,
		).
		Env(env...).
		Paths(path.Join(venvDir, "bin"))
}
