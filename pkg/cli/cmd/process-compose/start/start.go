package processcomposestart

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	pc "github.com/sdsc-ordes/quitsh/pkg/exec/process-compose"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/cobra"
)

const longDesc = `Start a process-compose definition from a 'devenv.sh' Nix shell
specified by an attribute path (e.g. 'mynamespace.shells.test-dbs') or installable
(e.g. './tools/nix#mynamespace.shells.test-dbs')
in a 'flake.nix' file.`

const timeoutWait = 100 * time.Second
const timeoutWaitInterval = 100 * time.Millisecond

type (
	startArgs struct {
		attrPath       string
		flakeDir       string
		socketPathFile string

		waitFor             []string
		waitForReady        []string
		attach              bool
		timeoutWait         time.Duration
		timeoutWaitInterval time.Duration
	}
)

func AddCmd(cl cli.ICLI, parent *cobra.Command, defaultFlakeDir string) {
	var stArgs startArgs

	startCmd := &cobra.Command{
		Use:     "start [devenv-attr-path or devenv-installable]",
		Short:   "Start a process-compose definition from a 'devenv.sh' Nix shell.",
		Long:    longDesc,
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			stArgs.attrPath = args[0]

			_, err := startProcessCompose(
				cl.RootDir(),
				stArgs.flakeDir,
				stArgs.attrPath,
				stArgs.waitFor,
				stArgs.waitForReady,
				stArgs.socketPathFile,
				stArgs.timeoutWait,
				stArgs.timeoutWaitInterval,
				stArgs.attach)

			return err
		},
	}

	startCmd.Flags().
		StringVarP(&stArgs.flakeDir,
			"flake-dir", "f", defaultFlakeDir, "The flake directory which contains a 'flake.nix' file.")

	startCmd.Flags().
		StringArrayVarP(&stArgs.waitFor,
			"wait-for", "w", nil, "Wait for these processes to be running.")

	startCmd.Flags().
		StringArrayVarP(&stArgs.waitForReady,
			"wait-for-ready", "r", nil, "Wait for these processes to be ready.")

	startCmd.Flags().
		StringVarP(&stArgs.socketPathFile,
			"socket-path-file", "s", "", "The file where the process-compose socket path is written to.")

	startCmd.Flags().
		BoolVarP(&stArgs.attach,
			"attach", "a", false, "If after start we attach to the process-compose instance.")

	startCmd.Flags().
		DurationVar(&stArgs.timeoutWait,
			"timeout", timeoutWait, "The max. timeout (e.g. `100s`) for waiting on processes.")

	startCmd.Flags().
		DurationVar(&stArgs.timeoutWaitInterval,
			"timeout-interval", timeoutWaitInterval, "The max. timeout interval (e.g. `100ms`) for polling processes.")

	parent.AddCommand(startCmd)
}

// startProcessCompose starts the process-compose services from `flake.nix` in `flakeDir`
// defined in the installable `devenvShellInstallable`.
// You can wait for the processes names to be running with `waitFor`.
func startProcessCompose(
	rootDir string,
	flakeDir string,
	devenvShellAttrPath string,
	waitForRunning []string,
	waitForReady []string,
	socketPathFile string,
	timeoutWait time.Duration,
	timeoutWaitInterval time.Duration,
	attach bool,
) (
	pcCtx pc.ProcessComposeCtx,
	err error,
) {
	if strings.Contains(devenvShellAttrPath, "#") {
		pcCtx, err = pc.StartFromInstallable(
			rootDir,
			devenvShellAttrPath,
			false,
		)
	} else {
		pcCtx, err = pc.Start(rootDir, flakeDir, devenvShellAttrPath, false)
	}
	if err != nil {
		return pcCtx, errors.AddContext(err, "could not start process-compose")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutWait)
	defer cancel()

	var conds []pc.ProcessCond
	for i := range waitForRunning {
		conds = append(conds, pc.ProcessCond{Name: waitForRunning[i], State: pc.ProcessRunning})
	}

	for i := range waitForReady {
		conds = append(conds, pc.ProcessCond{Name: waitForReady[i], State: pc.ProcessReady})
	}

	if len(conds) != 0 {
		log.Info("Wait for processes.",
			"ready", waitForReady,
			"running", waitForRunning,
			"timeout", timeoutWait,
			"interval", timeoutWaitInterval)
	}

	fulfilled, err := pcCtx.WaitTill(ctx, timeoutWaitInterval, conds...)
	if err != nil {
		return pcCtx, errors.AddContext(err, "failed to wait for processes")
	} else if !fulfilled {
		return pcCtx, errors.New("timed out while waiting for ready conditions on processes")
	}

	summary, err := pcCtx.Get("list", "-o", "json")
	if err != nil {
		return pcCtx, errors.AddContext(err, "could not get process summary")
	}
	log.Info("Processes status.", "summary", strings.ReplaceAll(summary, "\t", "  "))

	if socketPathFile != "" {
		err = os.WriteFile(socketPathFile, []byte(pcCtx.Socket()), fs.DefaultPermissionsFile)
		if err != nil {
			log.WarnE(err, "Could not write socket path to file '%s'.", socketPathFile)
		}
	}

	if attach {
		e := pcCtx.Check("attach")
		if e != nil {
			log.ErrorE(err, "Error occurred in attach.")
		}
	}

	log.Infof("Inspect processes with 'quitsh process-compose start -a ...'.")
	log.Infof("Stop processes with 'quitsh process-compose stop ...'.")

	return pcCtx, nil
}
