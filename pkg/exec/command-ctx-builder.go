package exec

import (
	"os"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type CmdContextBuilder struct {
	cmdCtx *CmdContext

	withPathSet bool
}

// NewCommandCtx returns a simple command context.
func NewCommandCtx(cwd string) *CmdContext {
	return NewCmdCtxBuilder().Cwd(cwd).Build()
}

// NewCmdCtxBuilder returns a builder to build a command context.
func NewCmdCtxBuilder() CmdContextBuilder {
	ctx := CmdContext{}

	return CmdContextBuilder{cmdCtx: &ctx}.NoQuiet().CredentialFilter(nil)
}

// Build finalizes the context.
func (c CmdContextBuilder) Build() *CmdContext {
	return c.cmdCtx
}

// Quiet disables pipeing stdout,stderr and logging the commands.
// Error reporting is not affected.
func (c CmdContextBuilder) Quiet() CmdContextBuilder {
	c.cmdCtx.pipeOutput = false
	c.cmdCtx.logCommand = false
	c.cmdCtx.captureError = true

	return c
}

// NoQuiet pipes stdout,stderr and logs the commands (default)
// Error reporting is not affected.
func (c CmdContextBuilder) NoQuiet() CmdContextBuilder {
	c.cmdCtx.pipeOutput = true
	c.cmdCtx.logCommand = true
	c.cmdCtx.captureError = false

	return c
}

// ExitCodeHandler set the exit code handler for this context.
func (c CmdContextBuilder) ExitCodeHandler(handler ExitCodeHandler) CmdContextBuilder {
	c.cmdCtx.exitCodeHandler = handler

	return c
}

// Paths adds `paths` to the `PATH` env. variable on the context.
// Also enables to lookup the command executable and resolve it
// to an absolute path before calling the `exec` stdlib module.
// This is needed since the `exec.LookPath` does consult the normal
// `os.Getenv("PATH")` which is not what we want.
func (c CmdContextBuilder) Paths(paths ...string) CmdContextBuilder {
	if c.cmdCtx.env == nil {
		// If the environment is not set, construct it.
		c.cmdCtx.env = os.Environ()
	}

	// Prepend the paths to `PATH` (if it exists)
	path := strings.Join(paths, string(os.PathListSeparator))
	pathCur := c.cmdCtx.env.FindIdx("PATH")

	if pathCur.Defined() {
		c.cmdCtx.env[pathCur.Idx()] = "PATH=" + path + string(os.PathListSeparator) + pathCur.Value
	} else {
		c.cmdCtx.env = append(c.cmdCtx.env, "PATH="+path)
	}

	c.withPathSet = true
	c.cmdCtx.enableLookPath = true

	return c
}

// Env adds environment variables to the command.
// NOTE: To set completely no env. variables you need to use
// `EnvEmpty()` and not `nil`.
func (c CmdContextBuilder) Env(env ...string) CmdContextBuilder {
	if c.withPathSet {
		log.Panic("you cannot set envs after 'WithPaths'")
	}

	// NOTE:
	// If we add here something and the default is set (.env == nil -> means use
	// os.Environ()) -> we need to add it back!
	if env != nil && c.cmdCtx.env == nil {
		c.cmdCtx.env = os.Environ()
	}

	c.cmdCtx.env = append(c.cmdCtx.env, env...)

	return c
}

func (c CmdContextBuilder) EnvRemove(key ...string) CmdContextBuilder {
	if c.withPathSet {
		log.Panic("you cannot set envs after 'WithPaths'")
	}

	// NOTE:
	// If we remove here something and the default is set (.env == nil -> means use
	// os.Environ()) -> we need to add it back!
	if key != nil && c.cmdCtx.env == nil {
		c.cmdCtx.env = os.Environ()
	}

	c.cmdCtx.env = c.cmdCtx.env.Remove(key...)

	return c
}

// EnvEmpty uses a completely empty environment.
func (c CmdContextBuilder) EnvEmpty() CmdContextBuilder {
	c.cmdCtx.env = []string{}

	return c
}

// BaseCmd sets the base command.
func (c CmdContextBuilder) BaseCmd(cmd string) CmdContextBuilder {
	c.cmdCtx.baseCmd = cmd

	return c
}

// BaseArgs adds arguments to the base command.
func (c CmdContextBuilder) BaseArgs(args ...string) CmdContextBuilder {
	c.cmdCtx.baseArgs = append(c.cmdCtx.baseArgs, args...)

	return c
}

// PrependCommand prepends a command `cmd` with) `args` infront of the command.
func (c CmdContextBuilder) PrependCommand(cmd string, args ...string) CmdContextBuilder {
	n := make([]string, 0, len(c.cmdCtx.baseArgs)+len(args))
	n = append(n, args...)
	n = append(n, c.cmdCtx.baseCmd)
	n = append(n, c.cmdCtx.baseArgs...)

	c.cmdCtx.baseCmd = cmd
	c.cmdCtx.baseArgs = n

	return c
}

// Cwd sets the working dir.
func (c CmdContextBuilder) Cwd(cwd string) CmdContextBuilder {
	c.cmdCtx.cwd = cwd

	return c
}

// CredentialFilter sets the argument credential filter. If `args` is `nil`,
// the default filter is set.
func (c CmdContextBuilder) CredentialFilter(args []string) CmdContextBuilder {
	if args == nil {
		c.cmdCtx.filterArgs = DefaultCredentialFilter
	} else {
		c.cmdCtx.filterArgs = NewCredentialFilter(args)
	}

	return c
}

// EnableCaptureError enables capturing the `stderr`.
// When `SetQuiet` is used this is always the case.
func (c CmdContextBuilder) EnableCaptureError() CmdContextBuilder {
	c.cmdCtx.captureError = true

	return c
}

// EnableEnvPrint enables printing the environment on failed commands.
func (c CmdContextBuilder) EnableEnvPrint() CmdContextBuilder {
	c.cmdCtx.enableEnvPrint = false

	return c
}
