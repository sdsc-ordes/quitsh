package yttx

import "github.com/sdsc-ordes/quitsh/pkg/exec"

// NewCtxBuilder returns a new YTT command context builder.
func NewCtxBuilder() exec.CmdContextBuilder {
	return exec.NewCmdCtxBuilder().BaseCmd("ytt")
}
