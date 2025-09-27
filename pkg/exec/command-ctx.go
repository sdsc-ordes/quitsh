package exec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	strs "github.com/sdsc-ordes/quitsh/pkg/common/strings"
	cerr "github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/env"
	"github.com/sdsc-ordes/quitsh/pkg/exec/lookpath"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

var EnableEnvPrint = false //nolint:gochecknoglobals // Allowed for CLI disabling.

type (
	ExitCodeHandler func(cmdError *CmdError) error

	// CmdContext defines the command context to execute commands.
	CmdContext struct {
		baseCmd  string
		baseArgs []string

		cwd string
		env env.EnvList

		// Captures the error and return it in the error.
		captureError bool

		// Enables printing the environment in errors.
		enableEnvPrint bool

		// Enables piping stdout/stderr.
		pipeOutput bool

		// Std input passed to the command.
		stdin io.Reader

		// Disable printing the commands.
		logCommand bool
		filterArgs ArgsFilter

		// The custom exit code handler for this context.
		// It handles all exit code (also 0):
		// Gets passed `nil` on exit code 0,
		// and a `CmdError` on any other error.
		exitCodeHandler ExitCodeHandler

		// Enables the resolution of the `baseCmd` (if not empty)
		// or the first argument before calling the command.
		// The stdlib `exec` module looks up the path with `os.Getenv("PATH)`
		// which is not an option, we would need to modify the `PATH` globally.
		enableLookPath bool
	}

	Waiter struct {
		c   *CmdContext
		cmd *exec.Cmd
		buf *bytes.Buffer
	}
)

// Cwd returns the working directory.
func (c *CmdContext) Cwd() string {
	return c.cwd
}

// BaseCmd returns the base command.
func (c *CmdContext) BaseCmd() string {
	return c.baseCmd
}

// BaseArgs returns the base arguments.
func (c *CmdContext) BaseArgs() []string {
	return common.CopySlice(c.baseArgs)
}

// Env returns the environment values.
func (c *CmdContext) Env() []string {
	return common.CopySlice(c.env)
}

// WithStdin sets a standard input reader to be used once.
func (c *CmdContext) WithStdin(r io.Reader) *CmdContext {
	c.stdin = r

	return c
}

// GetSplit executes a command and splits the output by newlines.
func (c *CmdContext) GetSplit(args ...string) ([]string, error) {
	return c.GetSplitWithEC(c.exitCodeHandler, args...)
}

// GetSplitWithEC executes a command with custom handling the exit code
// and splits the output by newlines.
func (c *CmdContext) GetSplitWithEC(handleExit ExitCodeHandler, args ...string) ([]string, error) {
	out, err := c.GetWithEC(handleExit, args...)
	if out == "" {
		return nil, err
	}

	return strs.SplitLines(out), err
}

// GetWithEC executes a command and gets the stdout (white-space trimmed).
// and custom handles the exit code.
// Returns `CmdError` on error.
func (c *CmdContext) GetWithEC(handleExit ExitCodeHandler, args ...string) (string, error) {
	baseCmd, args, err := c.getCommand(args)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.env

	buf := setupCapture(c, cmd, false)

	if c.logCommand {
		logCommand(c, cmd)
	}
	stdout, err := cmd.Output()
	err = handleExitCode(
		c, cmd, err, buf,
		c.enableEnvPrint || EnableEnvPrint, handleExit,
	)

	return strings.TrimSpace(string(stdout)), err
}

// GetStdErrWithEC executes a command and gets the stdout & stderr (white-space trimmed).
// and custom handles the exit code.
// Returns `CmdError` on error.
func (c *CmdContext) GetStdErrWithEC(
	handleExit ExitCodeHandler,
	args ...string,
) (string, string, error) {
	baseCmd, args, err := c.getCommand(args)
	if err != nil {
		return "", "", err
	}

	cmd := exec.Command(baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.env

	// Force capturing the stderr.
	buf := setupCapture(c, cmd, true)

	if c.logCommand {
		logCommand(c, cmd)
	}
	stdout, err := cmd.Output()

	b := buf
	if !c.captureError {
		// Set `buf` to `nil` to disable
		// setting the stderr again on the error.
		b = nil
	}
	err = handleExitCode(
		c, cmd, err, b,
		c.enableEnvPrint || EnableEnvPrint, handleExit,
	)

	return strings.TrimSpace(string(stdout)), strings.TrimSpace(buf.String()), err
}

// Get executes a command and gets the stdout (white-space trimmed).
// Returns `CmdError` on error.
func (c *CmdContext) Get(args ...string) (string, error) {
	return c.GetWithEC(c.exitCodeHandler, args...)
}

// GetStdErr executes a command and gets the stdout & stderr (white-space trimmed).
// Returns `CmdError` on error.
func (c *CmdContext) GetStdErr(args ...string) (string, string, error) {
	return c.GetStdErrWithEC(c.exitCodeHandler, args...)
}

// GetCombinedWithEC executes a command and gets the stdout and stderr (white-space trimmed).
// and custom handles the exit code.
// Returns `CmdError` on error.
func (c *CmdContext) GetCombinedWithEC(handleExit ExitCodeHandler, args ...string) (string, error) {
	baseCmd, args, err := c.getCommand(args)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.env

	if c.logCommand {
		logCommand(c, cmd)
	}
	stdout, err := cmd.CombinedOutput()
	err = handleExitCode(
		c, cmd, err, nil,
		c.enableEnvPrint || EnableEnvPrint, handleExit,
	)

	return strings.TrimSpace(string(stdout)), err
}

// GetCombined executes a command and gets the stdout and stderr (white-space trimmed).
// Returns `CmdError` on error.
func (c *CmdContext) GetCombined(args ...string) (string, error) {
	return c.GetCombinedWithEC(c.exitCodeHandler, args...)
}

// CheckWithEC executes a command and custom handles
// the exit code (will no use the internally set `exitCodeHandler`)
// Returns `CmdError` on error.
func (c *CmdContext) CheckWithEC(handleExit ExitCodeHandler, args ...string) error {
	baseCmd, args, err := c.getCommand(args)
	if err != nil {
		return err
	}

	cmd := exec.Command(baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.env

	if c.pipeOutput {
		cmd.Stdout = os.Stdout
	}

	buf := setupCapture(c, cmd, false)

	if c.logCommand {
		logCommand(c, cmd)
	}
	err = cmd.Run()

	return handleExitCode(c, cmd, err, buf, c.enableEnvPrint || EnableEnvPrint, handleExit)
}

// Check checks if a command executed successfully (optionally use `exitCodeHandler`)
// Returns `CmdError` on error.
func (c *CmdContext) Check(args ...string) error {
	return c.CheckWithEC(c.exitCodeHandler, args...)
}

// CheckPipe returns the stdout pipe like [exec.Cmd.StdoutPipe].
// [Waiter.Wait] must be called on `waiter` to wait for the error and must be called
// **first** when all read from the pipe have finished.
func (c *CmdContext) CheckPipe(args ...string) (waiter Waiter, pipe io.ReadCloser, err error) {
	baseCmd, args, err := c.getCommand(args)
	if err != nil {
		return
	}

	cmd := exec.Command(baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.env

	waiter.buf = setupCapture(c, cmd, false)

	pipe, err = cmd.StdoutPipe()
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	waiter.cmd = cmd
	waiter.c = c

	return
}

// Wait waits till the cmd has finished.
func (w Waiter) Wait() error {
	err := w.cmd.Wait()

	return handleExitCode(w.c, w.cmd, err, w.buf, w.c.enableEnvPrint || EnableEnvPrint, nil)
}

// WithTemplate renders a template to a temporary file and runs the
// `run` function on it.
func WithTemplate[T any](
	c *CmdContext,
	templ string,
	data any,
	run func(c *CmdContext, file string) (T, error),
) (res T, err error) {
	temp, err := template.New("1").Parse(templ)
	if err != nil {
		return
	}

	temp.Option("missingkey=error")

	f, err := os.CreateTemp("", "")
	if err != nil {
		return
	}
	defer f.Close()

	err = temp.Execute(f, data)
	if err != nil {
		return
	}

	return run(c, f.Name())
}

func formatError(cmd *exec.Cmd, stderr string, exitCode int, enableEnvPrint bool) string {
	e := cmd.Env
	if e == nil {
		e = []string{"... same as calling proc. ..."}
	} else if !enableEnvPrint {
		e = []string{"... disabled ..."}
	}

	return fmt.Sprintf(
		"command failed: '%q',\n  - cwd: '%s',\n  - env: %q,\n  - exit: %v,\n  - stderr:\n",
		cmd.Args,
		cmd.Dir,
		e,
		exitCode,
	) + strs.Indent(
		stderr,
		"  |  ",
	)
}

func (c *CmdContext) getCommand(args []string) (baseCmd string, argsOut []string, err error) {
	argsOut = append(argsOut, c.baseArgs...)

	if c.baseCmd != "" {
		baseCmd = c.baseCmd
		argsOut = append(argsOut, args...)
	} else {
		if len(args) < 1 {
			err = cerr.New("arguments do not contain any command")

			return
		}

		baseCmd = args[0]                      // Take first argument.
		argsOut = append(argsOut, args[1:]...) // Return the rest.
	}

	if c.enableLookPath && c.env != nil {
		if path := c.env.FindIdx("PATH"); path.Value != "" {
			lp, e := lookpath.Look(baseCmd, path.Value)
			if e == nil {
				baseCmd = lp
			}
		}
	}

	return
}

func setupCapture(c *CmdContext, cmd *exec.Cmd, forceCapture bool) (buf *bytes.Buffer) {
	if c.stdin != nil {
		cmd.Stdin = c.stdin
		c.stdin = nil // only use stdin exactly once
	}

	if c.captureError || forceCapture {
		buf = bytes.NewBuffer(nil)
		if c.pipeOutput {
			cmd.Stderr = io.MultiWriter(os.Stderr, buf)
		} else {
			cmd.Stderr = buf
		}
	} else if c.pipeOutput {
		cmd.Stderr = os.Stderr
	}

	return
}

func handleExitCode(
	c *CmdContext,
	cmd *exec.Cmd,
	err error,
	buf *bytes.Buffer,
	enableEnvPrint bool,
	exitCodeHandler ExitCodeHandler,
) error {
	if err == nil {
		if exitCodeHandler != nil {
			return wrapToNil(exitCodeHandler(nil))
		} else {
			return nil
		}
	}

	var exitErr *exec.ExitError
	exitCode := 255
	if t, ok := err.(*exec.ExitError); ok { //nolint:errorlint  // exec returns no wrapped error.
		exitErr = t
		exitCode = t.ExitCode()
	}

	// Extract stderr if possible.
	var stderr = "not captured"
	if buf != nil {
		stderr = buf.String()
	} else if log.IsDebug() && exitErr != nil {
		stderr = string(exitErr.Stderr)
	}

	if c.filterArgs != nil {
		cmd.Args = c.filterArgs(cmd.Args)
	}

	e := NewCmdError(cmd, stderr, exitCode, enableEnvPrint)

	if exitErr != nil {
		// Handle ExitError.
		if exitCodeHandler != nil {
			return exitCodeHandler(&e)
		} else {
			return e
		}
	} else {
		// Other reason. Wrap with original `err`.
		return cerr.Combine(e, err)
	}
}

// wrapToNil is a helper function which ensures returned values from `exitCodeHandler`
// return `nil`, if the `exitCodeHandler(nil)` is called and the implementer
// returns `nil` it is `<*CmdError, nil>` wide pointer which does not
// compare to `nil` !
func wrapToNil(e error) error {
	zero := (*CmdError)(nil)
	if errors.Is(e, zero) {
		return nil
	}

	return e
}

// logCommand will log the command and filter arguments (for security).
func logCommand(c *CmdContext, cmd *exec.Cmd) {
	var a = cmd.Args

	if c.filterArgs != nil {
		a = c.filterArgs(cmd.Args)
	}

	log.Debugf("Executing. cmd: '%q'", a)
}
