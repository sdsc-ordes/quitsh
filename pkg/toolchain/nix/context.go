package nixtoolchain

import (
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
)

// NewCtxBuilder, see `nix.NewDevShellCtxBuilder`.
func NewCtxBuilder(rootDir string, flakePath string, toolchain string) exec.CmdContextBuilder {
	installable := nix.ToolchainInstallable(flakePath, toolchain)

	return nix.NewDevShellCtxBuilderI(rootDir, installable)
}

// WrapOverToolchain, see `nix.WrapOverToolchain`.
func WrapOverToolchain(
	ctxBuilder exec.CmdContextBuilder,
	rootDir string,
	flakePath string,
	toolchain string,
) exec.CmdContextBuilder {
	if nix.HaveToolchain(toolchain) {
		return ctxBuilder
	}

	return nix.WrapOverDevShellI(ctxBuilder,
		rootDir,
		nix.ToolchainInstallable(flakePath, toolchain))
}
