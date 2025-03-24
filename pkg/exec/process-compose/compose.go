package processcompose

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// ProcessComposeCtx represents a `process-compose` context.
type ProcessComposeCtx struct {
	*exec.CmdContext
	socket string
}

// Start starts the process compose from a
// `devShellInstallable` (e.g. `./tools/nix#custodian.shells.test-dbs`
// which must be a `devenv` shell) in the flake located
// at `flakeDir`. The `rootDir` is the working directory and
// where the `.devenv/state/pwd` file is for `nonPureEval == false`.
func Start(
	rootDir string,
	flakeDir string,
	devShellInstallable string,
	logFile string,
) (pc ProcessComposeCtx, err error) {
	devShellInstallable = nix.FlakeInstallable(flakeDir, devShellInstallable)

	procCompExe, socketPath, err := getSocketPath(devShellInstallable, rootDir)
	if err != nil {
		return pc, err
	}

	err = os.MkdirAll(path.Dir(socketPath), fs.DefaultPermissionsDir)
	if err != nil {
		return pc, err
	}

	procfileScript, err := buildProcFileScript(devShellInstallable, rootDir)
	if err != nil {
		return pc, err
	}

	b := exec.NewCmdCtxBuilder().Cwd(rootDir)
	if logFile != "" {
		b = b.Env("PC_LOG_FILE=" + logFile)
	}

	// Start the process compose.
	err = b.Build().Check(procfileScript, "-D")
	if err != nil {
		return pc, errors.AddContext(err, "Could not start procfileScript '%s'.", procfileScript)
	}
	log.Info(
		"Launched process-compose for devenv shell.",
		"shell",
		devShellInstallable,
		"socket",
		socketPath,
	)

	b = exec.NewCmdCtxBuilder().
		Cwd(rootDir).
		BaseCmd(procCompExe).
		BaseArgs("--unix-socket", socketPath)

	return ProcessComposeCtx{
		CmdContext: b.Build(),
		socket:     socketPath}, nil
}

// Socket returns the socket used.
func (pc *ProcessComposeCtx) Socket() string {
	return pc.socket
}

// Stop stops the process compose.
func (pc *ProcessComposeCtx) Stop() error {
	return pc.Check("down")
}

// Check if processes in the process compose is running.
//
//nolint:gocognit // The goroutine polling is fairily simple to understand.
func (pc *ProcessComposeCtx) WaitTillRunning(
	ctx context.Context,
	procs ...string) (isRunning bool, err error) {
	err = pc.waitForSocket()
	if err != nil {
		return false, err
	}

	for {
		select {
		case <-ctx.Done():
			return false, nil
		default:
			var js string
			js, err = pc.Get("list", "-o", "json")
			if err != nil {
				return false, err
			}

			type Data struct {
				Status string `json:"status"`
				Name   string `json:"name"`
			}
			var d []Data

			err = json.Unmarshal([]byte(js), &d)
			if err != nil {
				return false,
					errors.AddContext(err, "Could not unmarshall output from process-compose.")
			}

			countRunning := 0
			for i := range d {
				if !slices.Contains(procs, d[i].Name) {
					continue
				}
				s := strings.ToLower(d[i].Status)

				switch s {
				case "completed":
					// If any is completed just report back.
					return false, nil
				case "running":
					countRunning += 1
				}
			}

			if countRunning == len(procs) {
				return true, nil
			}
		}
	}
}

func (pc *ProcessComposeCtx) waitForSocket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New("Process compose socket was not created in 10 seconds -> Timeout.")
		default:
			if fs.Exists(pc.Socket()) {
				return nil
			}
		}
	}
}

func getSocketPath(
	devShellInstallable string,
	rootDir string,
) (procCompExe string, socketPath string, err error) {
	nixx := nix.NewEvalCtx(rootDir)

	manager, err := nixx.Get("--raw", devShellInstallable+".config.process.manager.implementation")
	if err != nil {
		return
	}

	if manager != "process-compose" {
		err = errors.New("Only process-manager is supported in dev. shell: manager: '%v'", manager)

		return
	}

	procCompExe, err = nixx.Get(
		"--raw",
		devShellInstallable+".config.process.managers.process-compose.package.outPath")
	if err != nil {
		return
	}
	procCompExe = path.Join(procCompExe, "bin/process-compose")

	socketPath, err = nixx.Get(
		"--raw",
		devShellInstallable+".config.process.managers.process-compose.unixSocket.path")

	return procCompExe, socketPath, err
}

func buildProcFileScript(installable string, rootDir string) (string, error) {
	//nolint:lll
	// More options on the process managers are here:
	// https://github.com/cachix/devenv/blob/b2d2d5a20cfff742efb3c6dddbf47c66893b2d61/src/modules/process-managers/process-compose.nix#L96
	nixCtx := nix.NewBuildCtx(rootDir)

	procFileScript := installable + ".config.procfileScript"
	js, err := nixCtx.Get("--no-link", "--json", procFileScript)
	if err != nil {
		return "", err
	}

	type Out struct {
		Out string `json:"out"`
	}
	type Data struct {
		Outputs Out `json:"outputs"`
	}

	var d []Data

	err = json.Unmarshal([]byte(js), &d)
	if err != nil || len(d) == 0 {
		return "", errors.AddContext(err, "Could not unmarshal output from Nix '%v'.", js)
	}

	return d[0].Outputs.Out, nil
}
