package ssh

import (
	"os"
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
)

// StartAgent starts the `ssh-agent` if its not yet running and returns
// the socket path.
func StartAgent() (socketPath string, err error) {
	exeCtx := exec.NewCommandCtx(".")
	start := false

	err = exeCtx.CheckWithEC(func(cmdError *exec.CmdError) error {
		if cmdError != nil && cmdError.ExitCode() == 2 {
			start = true

			return nil
		}

		return cmdError
	}, "ssh-add", "-l")
	if err != nil {
		return "", err
	}

	if start {
		d, e := os.MkdirTemp("", "ssh-agent")
		if e != nil {
			return "", e
		}

		f := path.Join(d, "socket")

		return f, exeCtx.Check("ssh-agent", "-a", f)
	}

	socketPath = os.Getenv("SSH_AUTH_SOCK")
	if socketPath == "" {
		return "", errors.New("SSH_AUTH_SOCK is not defined, but should be since agent is running!")
	}

	return socketPath, nil
}
