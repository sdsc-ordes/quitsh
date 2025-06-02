package processcompose

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
	"golang.org/x/sync/errgroup"
)

// ProcessComposeCtx represents a `process-compose` context.
type ProcessComposeCtx struct {
	*exec.CmdContext
	socket string

	tempDir string
	logFile string
}

// Start starts the process compose from a
// `devShellAttrPath` (e.g. `custodian.shells.test-dbs`
// which must be a `devenv` shell) in the flake `flake.nix` located
// at `flakeDir`. The `rootDir` is the working directory and
// where the `.devenv/state/pwd` file is for `nonPureEval == false`.
// Note: You also call [StartFromInstallable] and directly pass an
// installable, e.g. a flake output attribute path like
// `./a/b/c#mynamespace.shells.test-dbs`.
func Start(
	rootDir string,
	flakeDir string,
	devShellAttrPath string,
	mustBeStarted bool,
) (pc ProcessComposeCtx, err error) {
	devShellAttrPath = nix.FlakeInstallable(flakeDir, devShellAttrPath)

	return StartFromInstallable(rootDir, devShellAttrPath, mustBeStarted)
}

// Start starts the process compose from a Nix
// `devShellInstallable` (e.g. `./tools/nix#custodian.shells.test-dbs`
// which must be a `devenv` shell).
// The `rootDir` is the working directory and
// where the `.devenv/state/pwd` file is for `nonPureEval == false`.
func StartFromInstallable(
	rootDir string,
	devShellInstallable string,
	mustBeStarted bool,
) (pc ProcessComposeCtx, err error) {
	procCompExe, socketPath, err := getSocketPath(devShellInstallable, rootDir)
	if err != nil {
		return pc, err
	}

	err = os.MkdirAll(path.Dir(socketPath), fs.DefaultPermissionsDir)
	if err != nil {
		return pc, err
	}

	procCompConfig, err := buildProcComposeConfigFile(devShellInstallable, rootDir)
	if err != nil {
		return pc, err
	}

	// Compute deterministic temp directory base on `procCompExe`.
	dir := path.Join(os.TempDir(),
		fmt.Sprintf("process-compose-%x",
			sha256.Sum256([]byte(procCompConfig))))
	err = os.MkdirAll(dir, fs.DefaultPermissionsDir)
	if err != nil {
		return pc, errors.AddContext(
			err,
			"could not create process-compose temp dir (logfile etc.).",
		)
	}
	logFile := path.Join(dir, "process-compose.log")

	b := exec.NewCmdCtxBuilder().
		Cwd(rootDir).
		BaseCmd(procCompExe).
		BaseArgs("--unix-socket", socketPath).
		Build()

	pc = ProcessComposeCtx{
		CmdContext: b,
		socket:     socketPath,
		tempDir:    dir,
		logFile:    logFile,
	}

	// Start the process compose.
	// Attach if the socket path does not exist
	// (the script already does it)
	if fs.Exists(socketPath) {
		log.Warnf("Socket '%s' is already existing. "+
			"Assume process-compose is started.", socketPath)

		return pc, nil
	} else {
		if mustBeStarted {
			return pc, errors.New("The process-compose instance must be started already but "+
				"socket '%s' is not existing.", socketPath)
		}

		log.Infof("Start process-compose with '%s'.", procCompConfig)
		err = b.Check(
			"--config", procCompConfig,
			"--keep-project",
			"--disable-dotenv",
			"--log-file", logFile,
			"--ordered-shutdown",
			"-D",
			"up")
		if err != nil {
			return pc, errors.AddContext(err, "Could not start process-compose with '%s'.", procCompConfig)
		}
	}

	log.Info(
		"Started process-compose for devenv shell.",
		"shell",
		devShellInstallable,
		"socket",
		socketPath,
		"logFile",
		logFile,
	)

	return pc, nil
}

// Socket returns the socket used.
func (pc *ProcessComposeCtx) Socket() string {
	return pc.socket
}

// LogFile returns the log file used.
func (pc *ProcessComposeCtx) LogFile() string {
	return pc.logFile
}

// Stop stops the process compose.
func (pc *ProcessComposeCtx) Stop() error {
	// Just forcefully delete the socket path and temp dir.
	defer func() {
		os.Remove(pc.socket)
		os.RemoveAll(pc.tempDir)
	}()

	return pc.Check("down")
}

// Check if processes in the process compose is running.
//
//nolint:gocognit // The goroutine polling is fairily simple to understand.
func (pc *ProcessComposeCtx) WaitTillRunning(
	ctx context.Context,
	procs ...string) (isRunning bool, err error) {
	if len(procs) == 0 {
		return true, nil
	}

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

	var manager string
	var pcPath string

	g := new(errgroup.Group)

	// Get manager.
	g.Go(func() error {
		val, e := nixx.Get("--raw", devShellInstallable+".config.process.manager.implementation")
		manager = val

		return e
	})

	// Get process-compose path.
	g.Go(func() error {
		val, e := nixx.Get(
			"--raw",
			devShellInstallable+".config.process.managers.process-compose.package.outPath",
		)
		pcPath = val

		return e
	})

	// Get socket path.
	g.Go(func() error {
		val, e := nixx.Get(
			"--raw",
			devShellInstallable+".config.process.managers.process-compose.unixSocket.path",
		)
		socketPath = val

		return e
	})

	// Wait for all goroutines
	if err = g.Wait(); err != nil {
		return "", "", err
	}

	if manager != "process-compose" {
		return "", "", errors.New(
			"Only process-manager is supported in dev. shell: manager: '%v'",
			manager,
		)
	}

	procCompExe = path.Join(pcPath, "bin/process-compose")

	return procCompExe, socketPath, err
}

func buildProcComposeConfigFile(installable string, rootDir string) (string, error) {
	//nolint:lll
	// More options on the process managers are here:
	// https://github.com/cachix/devenv/blob/b2d2d5a20cfff742efb3c6dddbf47c66893b2d61/src/modules/process-managers/process-compose.nix#L96
	// Devenv start the stuff on the attribute `.config.procfileScript` which we do not use.

	nixCtx := nix.NewBuildCtx(rootDir)

	configFile := installable + ".config.process.managers.process-compose.configFile"
	js, err := nixCtx.Get("--no-link", "--json", configFile)
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
