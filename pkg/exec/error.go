package exec

import (
	"os/exec"
)

type CmdError struct {
	exitCode int
	stderr   string
	message  string
}

func (c CmdError) Error() string {
	return c.message
}

func (c *CmdError) Stderr() string {
	return c.stderr
}

func (c *CmdError) ExitCode() int {
	return c.exitCode
}

// NewCmdError returns a [CmdError].
func NewCmdError(
	cmd *exec.Cmd,
	stderr string,
	exitCode int,
	enableEnvPrint bool,
) CmdError {
	return CmdError{
		message:  formatError(cmd, stderr, exitCode, enableEnvPrint),
		exitCode: exitCode,
		stderr:   stderr,
	}
}
