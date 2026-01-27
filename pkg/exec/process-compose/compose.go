package processcompose

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"golang.org/x/sync/errgroup"
)

const ProcessRunning ProcessState = 0
const ProcessReady ProcessState = 1
const ProcessCompleted ProcessState = 2

// ProcessComposeCtx represents a `process-compose` context.
type (
	ProcessComposeCtx struct {
		*exec.CmdContext
		socket string

		tempDir string
		logFile string
	}

	ProcessState int

	ProcessCond struct {
		Name  string
		State ProcessState
	}
)

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

// StartFromInstallable starts the process compose from a Nix
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

	// We need to launch the process-compose over a
	// devShell to start it properly.
	build := func(b exec.CmdContextBuilder) *exec.CmdContext {
		return b.
			Cwd(rootDir).
			BaseArgs("--unix-socket", socketPath).
			Build()
	}
	pcCtxDev := build(nix.NewDevShellCtxBuilderI(
		rootDir, devShellInstallable).BaseArgs(procCompExe))

	pc = ProcessComposeCtx{
		CmdContext: build(exec.NewCmdCtxBuilder().BaseCmd(procCompExe)),
		socket:     socketPath,
		tempDir:    dir,
		logFile:    logFile,
	}

	log.Info("Settings for process-compose.",
		"rootDir", rootDir,
		"installable", devShellInstallable,
		"procCompExe", procCompExe,
		"config", procCompConfig,
		"socketPath", socketPath,
		"logFile", logFile)

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
		log.Info("Start process-compose.")

		err = pcCtxDev.Check(
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

// WaitTill checks if processes in the process compose is running.
//
//nolint:gocognit // The goroutine polling is fairily simple to understand.
func (pc *ProcessComposeCtx) WaitTill(
	ctx context.Context,
	checkInterval time.Duration,
	conds ...ProcessCond) (fulfilled bool, err error) {
	if len(conds) == 0 {
		return true, nil
	}

	err = pc.waitForSocket()
	if err != nil {
		return false, err
	}

	type ProcInfo struct {
		Status  string `json:"status"`
		IsReady string `json:"is_ready"` //nolint:tagliatelle // external input
		Name    string `json:"name"`
	}

	// Map to keep track of logged statues of procs.
	reported := make(map[string]ProcInfo)

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

			var procs []ProcInfo

			err = json.Unmarshal([]byte(js), &procs)
			if err != nil {
				return false,
					errors.AddContext(err,
						"Could not unmarshall output from process-compose.\n'%s'",
						js)
			}

			condsFulfilled := 0

			for j := range procs {
				p := &procs[j]
				// All lowercase, to be safe.
				p.Status = strings.ToLower(p.Status)
				p.IsReady = strings.ToLower(p.IsReady)

				if reported[p.Name].Status != p.Status {
					log.Infof(
						"Process status change: '%s': '%s' -> '%s'.",
						p.Name,
						reported[p.Name].Status,
						p.Status,
					)
				}
				if reported[p.Name].IsReady != p.IsReady {
					log.Infof(
						"Process readiness change: '%s': '%s' -> '%s'.",
						p.Name,
						reported[p.Name].IsReady,
						p.IsReady,
					)
				}
				reported[p.Name] = *p

				for i := range conds {
					cond := &conds[i]
					if cond.Name != p.Name {
						continue
					}

					switch {
					case cond.State == ProcessRunning && p.Status == "running":
						log.Infof("Process condition: '%s': 'running' ✅", p.Name)

						fallthrough
					case cond.State == ProcessReady && p.IsReady == "ready":
						log.Infof("Process condition: '%s': 'ready' ✅", p.Name)

						fallthrough
					case cond.State == ProcessCompleted && p.Status == "completed":
						log.Infof("Process condition: '%s': 'completed' ✅", p.Name)
						condsFulfilled += 1
					}
				}
			}

			if condsFulfilled == len(conds) {
				return true, nil
			}
		}

		// Sleep for or until context is cancelled or check interval reached.
		select {
		case <-ctx.Done():
			return false, nil
		case <-time.After(checkInterval):
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
