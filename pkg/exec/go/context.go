package gox

import "github.com/sdsc-ordes/quitsh/pkg/exec"

// NewCtxBuilder returns a new Go command context builder.
func NewCtxBuilder() exec.CmdContextBuilder {
	return exec.NewCmdCtxBuilder().BaseCmd("go")
}
