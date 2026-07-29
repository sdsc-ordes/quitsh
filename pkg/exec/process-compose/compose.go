package processcompose

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	stderr "errors"
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

const ProcessComposeOverDevenv ProcessComposeImpl = "devenv"
const ProcessComposeOverServicesFlake ProcessComposeImpl = "process-compose-flake"

// ProcessComposeCtx represents a `process-compose` context.
type (
	// ProcessComposeImpl represents the implementation
	// if the process-compopse is defined in a `devenv` NixShell or
	// over `process-compose-flake`.
	ProcessComposeImpl string

	ProcessComposeCtx struct {
		exec.CmdContext
		version string
		socket  string

		startCmd []string

		tempDir string
		logFile string

		log log.ILog
	}

	ProcessState int

	ProcessCond struct {
		Name  string
		State ProcessState

		// Fail if the state becomes completed.
		// Is `false` by default.
		NoFailOnCompleted bool
	}

	StartOption func(*startOpts) error

	startOpts struct {
		socketPathFile string
		mustBeStarted  bool
	}
)

// Start starts the process compose from a
// `devShellAttrPath` (e.g. `custodian.shells.test-dbs`
// which must be a `devenv` shell) in the flake `flake.nix` located
// at `flakeDir`. The `rootDir` is the working directory and
// where the `.devenv/state/pwd` file is for `nonPureEval == false`.
// Note: You also call [StartFromInstallable] and directly pass an
// installable, e.g. a flake output attribute path like
// `./a/b/c#mynamespace.test-dbs`.
func Start(
	log log.ILog,
	rootDir string,
	flakeDir string,
	attrPath string,
	impl ProcessComposeImpl,
	opts ...StartOption,
) (pc *ProcessComposeCtx, err error) {
	attrPath = nix.FlakeInstallable(flakeDir, attrPath)

	return StartFromInstallable(log, rootDir, attrPath, impl, opts...)
}

// StartFromInstallable starts the process compose from a Nix
// `installable` (e.g. `./tools/nix#custodian.shells.test-dbs`
// which must be a `devenv` shell or a `process-compose-flake` derivation).
// The `rootDir` is the working directory and
// where the `.devenv/state/pwd` file is for `nonPureEval == false`.
func StartFromInstallable(
	log log.ILog,
	rootDir string,
	installable string,
	impl ProcessComposeImpl,
	opts ...StartOption,
) (pc *ProcessComposeCtx, err error) {
	var o startOpts
	err = o.Apply(opts...)
	if err != nil {
		return nil, err
	}

	var pcCtxDev *exec.CmdContext

	switch impl {
	case ProcessComposeOverDevenv:
		pc, pcCtxDev, err = settingsFromDevenv(log, rootDir, installable)
	case ProcessComposeOverServicesFlake:
		pc, pcCtxDev, err = settingsFromServicesFlake(log, rootDir, installable)
	default:
		return nil, errors.New(
			"Implementation '%v' for process-compose start is not supported.",
			impl,
		)
	}

	if err != nil {
		return nil, errors.AddContext(err, "Could not create process-compose ctx.")
	}

	// Write the socket path to the file
	if o.socketPathFile != "" {
		log.Infof("Write socket path file '%s'.", o.socketPathFile)
		err = os.WriteFile(o.socketPathFile, []byte(pc.Socket()), fs.DefaultPermissionsFile)
		if err != nil {
			return pc, errors.AddContext(
				err,
				"Could not write socket path to file '%s'.",
				o.socketPathFile,
			)
		}
	}

	// Start the process compose.
	// Attach if the socket path does not exist
	// (the script already does it)
	if fs.Exists(pc.Socket()) {
		log.Warnf("Socket '%s' is already existing. "+
			"Assume process-compose is started.", pc.Socket())

		return pc, nil
	} else {
		if o.mustBeStarted {
			return pc, errors.New("The process-compose instance must be started already but "+
				"socket '%s' is not existing.", pc.Socket())
		}
		log.Info("Start process-compose.")

		err = pcCtxDev.Check(pc.startCmd...)
		if err != nil {
			return pc, errors.AddContext(err, "Could not start process-compose with start command '%v'.", pc.startCmd)
		}
	}

	log.Info(
		"Started process-compose.",
		"impl",
		impl,
		"installable",
		installable,
		"socket",
		pc.Socket(),
		"logFile",
		pc.LogFile(),
	)

	return pc, nil
}

func settingsFromServicesFlake(
	log log.ILog,
	rootDir, installable string,
) (*ProcessComposeCtx, *exec.CmdContext, error) {
	log.Infof("Getting settings for 'process-compose-flake' implementation.")

	// Compute deterministic temp directory base on `procCompExe`.
	dir := path.Join(os.TempDir(), "process-compose",
		"process-compose-"+fmt.Sprintf("%x",
			sha256.Sum256([]byte(installable)))[:6])

	err := os.MkdirAll(dir, fs.DefaultPermissionsDir)
	if err != nil {
		return nil, nil, errors.AddContext(
			err,
			"could not create process-compose temp dir (logfile etc.).",
		)
	}
	logFile := path.Join(dir, "process-compose.log")

	socketPath := path.Join(dir, "pc.sock")

	startCmd := []string{
		"--keep-project",
		"--disable-dotenv",
		"--ordered-shutdown", //nolint:goconst // ok here.
		"--log-file", logFile,
		"--no-server=false",
		"-D",
		"up"}

	pc := nix.NewRunCtx(
		rootDir,
		func(b exec.CmdContextBuilder) exec.CmdContextBuilder {
			return b.Cwd(rootDir).BaseArgs(installable, "--",
				"--unix-socket", socketPath)
		},
	)

	version := "nix"

	log.Info("Settings for process-compose.",
		"version", version,
		"rootDir", rootDir,
		"installable", installable,
		"socketPath", socketPath,
		"logFile", logFile)

	pcCtx := &ProcessComposeCtx{
		CmdContext: *pc.CmdContext,
		version:    version,
		log:        log,
		socket:     socketPath,
		startCmd:   startCmd,
		tempDir:    dir,
		logFile:    logFile,
	}

	return pcCtx, &pcCtx.CmdContext, nil
}

func settingsFromDevenv(
	log log.ILog,
	rootDir, installable string,
) (*ProcessComposeCtx, *exec.CmdContext, error) {
	log.Infof("Getting settings for 'devenv' implementation.")
	procCompExe, socketPath, err := getSocketPath(installable, rootDir)
	if err != nil {
		return nil, nil, err
	}

	err = os.MkdirAll(path.Dir(socketPath), fs.DefaultPermissionsDir)
	if err != nil {
		return nil, nil, err
	}

	procCompConfig, err := buildProcComposeConfigFile(installable, rootDir)
	if err != nil {
		return nil, nil, err
	}

	// Compute deterministic temp directory base on `installable`.
	dir := path.Join(os.TempDir(), "process-compose",
		"process-compose-"+fmt.Sprintf("%x",
			sha256.Sum256([]byte(installable)))[:6])

	err = os.MkdirAll(dir, fs.DefaultPermissionsDir)
	if err != nil {
		return nil, nil, errors.AddContext(
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
		rootDir, installable).BaseArgs(procCompExe))

	pcB := exec.NewCmdCtxBuilder().BaseCmd(procCompExe)
	version, err := pcB.Build().Get("version")
	if err != nil {
		return nil, nil, err
	}

	log.Info("Settings for process-compose.",
		"version", version,
		"rootDir", rootDir,
		"installable", installable,
		"procCompExe", procCompExe,
		"config", procCompConfig,
		"socketPath", socketPath,
		"logFile", logFile)

	startCmd := []string{
		"--config", procCompConfig,
		"--keep-project",
		"--disable-dotenv",
		"--ordered-shutdown",
		"--log-file", logFile,
		"--ordered-shutdown",
		"-D",
		"up"}

	return &ProcessComposeCtx{
		CmdContext: *build(pcB),
		version:    version,
		log:        log,
		socket:     socketPath,
		startCmd:   startCmd,
		tempDir:    dir,
		logFile:    logFile,
	}, pcCtxDev, nil
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

type monitor struct {
	eventCh   <-chan pcEvent
	terminate func() error
}

func startPCMonitor(
	log log.ILog,
	pc *exec.CmdContext,
) (m monitor, err error) {
	eventCh := make(chan pcEvent, 10) //nolint:mnd
	m.eventCh = eventCh

	var c context.Context
	var cancel context.CancelCauseFunc
	var waiter exec.Waiter

	// Add a cancelation.
	pc = pc.ToBuilder().ContextWrap(
		func(p context.Context) context.Context {
			c, cancel = context.WithCancelCause(p) //nolint: fatcontext // false-positive

			return c
		}).Build()

	waiter, pipe, err := pc.CheckPipe("process", "monitor", "-o", "json")
	if err != nil {
		return m, errors.AddContext(err, "failed to start pipe")
	}

	// The terminate function to correctly handle errors.
	m.terminate = func() error {
		// Cancel the process. Note: errors are only set once when `cancel` called!
		normalTerminate := errors.New("normal terminate")
		cancel(normalTerminate)

		pipe.Close()

		e := waiter.Wait()
		cerr := c.Err()
		cause := context.Cause(c)
		switch {
		case stderr.Is(cause, normalTerminate):
			return nil
		case cause != nil:
			// A problem is to report.
			return cause
		case cerr != nil:
			return nil //nolint:nilerr // The context was apparently cancelled so nothing to report.
		}

		// ...otherwise report why process monitor failed.
		return e
	}

	// Read output in separate Go routine.
	go func() {
		defer func() {
			pipe.Close()
			close(eventCh)
			log.Info("Monitor reader for 'process-compose' closed.")
		}()

		var count int
		var event pcEvent
		s := bufio.NewScanner(pipe)
		for s.Scan() {
			line := s.Bytes()
			log.Tracef("Parse event '%s'", string(line))

			e := json.Unmarshal(line, &event)
			if e != nil {
				log.ErrorE(e, "Could not serialize process-compose monitor event: '%v', abort.",
					string(line))

				cancel(errors.AddContext(e,
					"could not serialize process-compose monitor event"))

				return
			}

			// Send event.
			// Note: if `ch` might be full we dont block since first case.
			select {
			case <-c.Done():
				log.Infof("Monitor context cancelled.")

				return
			case eventCh <- event:
				count++
			}
		}

		if e := s.Err(); e != nil {
			cancel(errors.AddContext(e, "event scanner failed"))
		}

		log.Infof("Monitor reader finished, processed '%v' events.", count)
	}()

	return m, nil
}

// WaitTill checks if processes in the process compose are running.
//
//nolint:gocognit
func (pc *ProcessComposeCtx) WaitTill(
	ctx context.Context,
	log log.ILog,
	conds ...ProcessCond) (fulfilled bool, err error) {
	if len(conds) == 0 {
		return true, nil
	}

	err = pc.waitForSocket()
	if err != nil {
		return false, err
	}

	// Map to keep track of all procs.
	condsFulfilled := 0
	cache := make(map[string]pcState)

	mon, err := startPCMonitor(log, &pc.CmdContext)
	if err != nil {
		return false, err
	}
	defer func() {
		log.Info("Terminating 'process-compose' monitor.")
		if e := mon.terminate(); e != nil {
			err = errors.Combine(
				errors.AddContext(e, "process monitor terminated with problems"),
				err,
			)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Error("WaitTill context closed. Conditions could not be fulfilled.")

			return false, nil
		case event, ok := <-mon.eventCh:
			if !ok {
				log.Info("Event channel closed.")

				return false, nil
			}

			log.Debug("Event received.", "event", event)
			p := &event.State

			// All lowercase, to be safe.
			p.Status = strings.ToLower(p.Status)
			p.IsReady = strings.ToLower(p.IsReady)

			if cache[p.Name].Status != p.Status {
				log.Infof(
					"Process status change: '%s': '%s' -> '%s'.",
					p.Name,
					cache[p.Name].Status,
					p.Status,
				)
			}
			if cache[p.Name].IsReady != p.IsReady {
				log.Infof(
					"Process readiness change: '%s': '%s' -> '%s'.",
					p.Name,
					cache[p.Name].IsReady,
					p.IsReady,
				)
			}
			cache[p.Name] = *p

			abort := evalConditions(log, conds, p, &condsFulfilled)
			if abort {
				return false, nil
			}

			if condsFulfilled == len(conds) {
				log.Infof("All conditions fulfilled.")

				return true, nil
			}

			log.Warnf("Not all conditions fulfilled: '%v/%v'",
				condsFulfilled, len(conds))
		}
	}
}

func evalConditions(
	log log.ILog,
	conds []ProcessCond,
	p *pcState,
	condsFulfilled *int,
) (abort bool) {
	for i := range conds {
		cond := &conds[i]
		if cond.Name != p.Name {
			continue
		}

		switch {
		case !cond.NoFailOnCompleted && p.Status == "completed":
			log.Warnf(
				"Process condition: '%s': 'completed' which is a guard and must not happen ❌.",
				p.Name,
			)

			return true
		case cond.State == ProcessRunning && p.Status == "running":
			log.Infof("Process condition: '%s': 'running' fulfilled ✅.", p.Name)
			*condsFulfilled += 1
		case cond.State == ProcessReady && p.IsReady == "ready":
			log.Infof("Process condition: '%s': 'ready' fulfilled ✅.", p.Name)
			*condsFulfilled += 1
		case cond.State == ProcessCompleted && p.Status == "completed":
			// FIXME: Set that to completed once.
			log.Error("Completed condition is not supported at the moment due to: " +
				"https://github.com/cachix/devenv/issues/2879")

			return true
			// Uncomment: for fixme above.
			// log.Infof("Process condition: '%s': 'completed' ✅", p.Name)
			// condsFulfilled += 1
		}
	}

	return false
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
	nixBuildx := nix.NewBuildCtx(rootDir)

	var manager string
	var pcPath string

	g := new(errgroup.Group)

	// Get manager.
	g.Go(func() error {
		val, e := nixx.Get("--raw", devShellInstallable+".config.process.manager.implementation")
		manager = val

		return e
	})

	// Build process-compose path.
	g.Go(func() error {
		drv, e := nixBuildx.BuildInstallable(
			devShellInstallable + ".config.process.managers.process-compose.package",
		)
		if e != nil {
			return e
		}

		pcPath = drv.Outputs.Out

		return nil
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

// SetLog sets the logger on the context.
func (pc *ProcessComposeCtx) SetLog(log log.ILog) {
	pc.log = log
}

// Apply applies all start options.
func (c *startOpts) Apply(options ...StartOption) error {
	for _, f := range options {
		if err := f(c); err != nil {
			return err
		}
	}

	return nil
}

// WithMustBeStarted only checks that the process-compose
// is started on the socket and does not start it.
func WithMustBeStarted(value bool) StartOption {
	return func(o *startOpts) error {
		o.mustBeStarted = value

		return nil
	}
}

// WithSocketPathFile writes the socket path to the file `file`.
func WithSocketPathFile(file string) StartOption {
	return func(o *startOpts) error {
		o.socketPathFile = file

		return nil
	}
}
