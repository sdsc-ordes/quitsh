package ssh

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type Agent struct {
	SocketPath string
	pid        int
}

type StartOption func(*opts)

type opts struct {
	newInstance bool
}

// StartAgent starts the `ssh-agent` if its not yet running and returns
// the socket path.
func StartAgent(log log.ILog, options ...StartOption) (agent Agent, err error) {
	var o opts
	o.Apply(options...)

	exeCtx := exec.NewCommandCtx(".")
	newInstance := o.newInstance

	if !newInstance {
		// Check if already running.
		err = exeCtx.CheckWithEC(func(cmdError *exec.CmdError) error {
			if cmdError != nil && cmdError.ExitCode() == 2 {
				log.Info("SSH agent is not running.")

				// Not running start a new one.
				newInstance = true

				return nil
			}

			return cmdError
		}, "ssh-add", "-l")
		if err != nil {
			return Agent{}, err
		}
	}

	if newInstance {
		log.Info("Start a new SSH agent.")

		d, e := os.MkdirTemp("", "ssh-agent")
		if e != nil {
			return Agent{}, errors.AddContext(e, "Could not start ssh-agent.")
		}

		f := path.Join(d, "socket")

		out, e := exeCtx.Get("ssh-agent", "-a", f)
		if e != nil {
			return Agent{},
				errors.AddContext(e, "Could not start ssh-agent.")
		}
		m := regexp.MustCompile(`SSH_AGENT_PID=(\d+)`).FindStringSubmatch(out)
		if len(m) < 2 { //nolint:mnd
			return Agent{},
				errors.New("Could not extract PID from ssh-agent start.")
		}

		pid, e := strconv.Atoi(m[1])
		if e != nil || pid == 0 {
			return Agent{},
				errors.AddContext(e, "Could not extract PID from ssh-agent start.")
		}

		log.Infof("Started SSH agent on socket '%v' and PID '%v'", f, pid)

		return Agent{SocketPath: f, pid: pid}, nil
	}

	socketPath := os.Getenv("SSH_AUTH_SOCK")
	if socketPath == "" {
		return Agent{},
			errors.New("SSH_AUTH_SOCK is not defined, but should be since agent is running!")
	}

	return Agent{socketPath, 0}, nil
}

// Close closes the agent if started.
func (c *Agent) Close() error {
	if c.pid == 0 {
		return nil
	}

	p, _ := os.FindProcess(c.pid)
	if p == nil {
		return nil
	}

	e := p.Kill()
	if e != nil {
		return errors.AddContext(e, "Could not kill ssh-agent with PID '%v'.", c.pid)
	}

	return nil
}

// Env returns the env. variables of the agent.
// If not started by its own, `SSH_AGENT_PID` is not returned.
func (s *Agent) Env() (e []string) {
	e = append(e, "SSH_AUTH_SOCK="+s.SocketPath)
	if s.pid != 0 {
		e = append(e, fmt.Sprintf("SSH_AGENT_PID=%v", s.pid))
	}

	return e
}

func (c *opts) Apply(options ...StartOption) {
	for _, f := range options {
		f(c)
	}
}

// WithNewInstance recreats a new agent instead of testing if one already runs.
func WithNewInstance(enable bool) StartOption {
	return func(o *opts) {
		o.newInstance = enable
	}
}
