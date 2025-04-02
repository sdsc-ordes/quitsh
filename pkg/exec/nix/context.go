package nix

import (
	"os"
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// NewCtxBuilder returns a new Nix command context builder.
func NewCtxBuilder() exec.CmdContextBuilder {
	return exec.NewCmdCtxBuilder().BaseCmd("nix")
}

// NewBuildCtxBuilder returns a new `nix build` command context for flake builds.
func NewBuildCtx(
	rootDir string,
	applyOpts ...func(b exec.CmdContextBuilder) exec.CmdContextBuilder) BuildCtx {
	c := addDefaultArgs(rootDir, exec.NewCmdCtxBuilder().BaseCmd("nix").BaseArgs("build"))
	for i := range applyOpts {
		c = applyOpts[i](c)
	}

	return BuildCtx{c.Build()}
}

type BuildCtx struct {
	*exec.CmdContext
}

// NewEvalCtxBuilder returns a new `nix eval` command context for flake evaluations.
func NewEvalCtx(
	rootDir string,
	applyOpts ...func(b exec.CmdContextBuilder) exec.CmdContextBuilder,
) NixEvalCtx {
	c := addDefaultArgs(
		rootDir, exec.NewCmdCtxBuilder().BaseCmd("nix").BaseArgs("eval"),
	)
	for i := range applyOpts {
		c = applyOpts[i](c)
	}

	return NixEvalCtx{c.Build()}
}

type NixEvalCtx struct {
	*exec.CmdContext
}

// NewEvalCtxBuilder returns a new `nix run` command context builder.
func NewRunCtx(
	rootDir string,
	applyOpts ...func(b exec.CmdContextBuilder) exec.CmdContextBuilder,
) NixRunCtx {
	c := addDefaultArgs(
		rootDir, exec.NewCmdCtxBuilder().BaseCmd("nix").BaseArgs("run"),
	)
	for i := range applyOpts {
		c = applyOpts[i](c)
	}

	return NixRunCtx{c.Build()}
}

type NixRunCtx struct {
	*exec.CmdContext
}

// NewDevShellCtxBuilder returns a new command context builder which runs all
// commands over a Nix development shell.
func NewDevShellCtxBuilder(
	rootDir string,
	flakePath string,
	attrPath string,
) exec.CmdContextBuilder {
	return NewDevShellCtxBuilderI(rootDir, FlakeInstallable(flakePath, attrPath))
}

func addDefaultArgs(rootDir string, b exec.CmdContextBuilder) exec.CmdContextBuilder {
	devenvRoot := rootDir + "/.devenv/state/pwd"

	err := os.MkdirAll(path.Dir(devenvRoot), fs.DefaultPermissionsDir)
	err = errors.Combine(err, os.WriteFile(devenvRoot, []byte(rootDir), fs.DefaultPermissionsFile))
	log.PanicE(err, "Devenv root file could not be written.", "path", devenvRoot)

	// We inject `--override-input` to set the `devenv-root` flake input (HACK).
	// This is currently needed for devenv to properly run in pure hermetic
	// mode while still being able to run processes & services and modify
	// (some parts) of the active shell.mkdir -p .devenv/state
	// See: https://github.com/cachix/devenv/issues/1461
	// NOTE: This will also work if no input matches the override which is important
	//       for users not using this.
	return b.
		Cwd(rootDir).
		BaseArgs(
			"--override-input",
			"devenv-root",
			"path:"+rootDir+"/.devenv/state/pwd",
			"--accept-flake-config")
}

// NewDevShellCtxBuilderI, see `NewDevShellCtxBuilder`.
func NewDevShellCtxBuilderI(rootDir string, installable string) exec.CmdContextBuilder {
	debug.Assert(path.IsAbs(rootDir), "Devenv root must be an absolute path.")

	return addDefaultArgs(
		rootDir,
		exec.NewCmdCtxBuilder().
			BaseCmd("nix").BaseArgs("develop"),
	).BaseArgs(installable, "--command")
}

// WrapOverDevShell wraps a command context builder
// over a dev shell with `NewDevShellCtxBuilder`.
func WrapOverDevShell(
	ctxBuilder exec.CmdContextBuilder,
	rootDir string,
	flakePath string,
	attrPath string,
) exec.CmdContextBuilder {
	return WrapOverDevShellI(ctxBuilder, rootDir, FlakeInstallable(flakePath, attrPath))
}

// WrapOverDevShellI, see `WrapOverDevShellI`.
func WrapOverDevShellI(
	ctxBuilder exec.CmdContextBuilder,
	rootDir string,
	installable string,
) exec.CmdContextBuilder {
	devShell := NewDevShellCtxBuilderI(rootDir, installable).Build()

	return ctxBuilder.PrependCommand(devShell.BaseCmd(), devShell.BaseArgs()...)
}
