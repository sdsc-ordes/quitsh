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

const timeoutContainers = 100 * time.Second

type startArgs struct {
	attrPath     string
	flakeDir     string
	waitFor      []string
	waitForReady []string

	attach         bool
	socketPathFile string
}

func AddCmd(cl cli.ICLI, parent *cobra.Command, defaultFlakeDir string) {
	var stArgs startArgs

	startCmd := &cobra.Command{
		Use:     "start [devenv-attr-path or devenv-installable]",
		Short:   "Start a process-compose definition from a 'devenv.sh' Nix shell.",
		Long:    longDesc,
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			stArgs.attrPath = args[0]

			_, err := StartServices(
				cl.RootDir(),
				stArgs.flakeDir,
				stArgs.attrPath,
				stArgs.waitFor,
				stArgs.waitForReady,
				stArgs.socketPathFile,
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

	parent.AddCommand(startCmd)
}

// StartServices starts the process-compose services from `flake.nix` in `flakeDir`
// defined in the installable `devenvShellInstallable`.
// You can wait for the processes names to be running with `waitFor`.
func StartServices(
	rootDir string,
	flakeDir string,
	devenvShellAttrPath string,
	waitForRunning []string,
	waitForReady []string,
	socketPathFile string,
	attach bool) (
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

	ctx, cancel := context.WithTimeout(context.Background(), timeoutContainers)
	defer cancel()

	var conds []pc.ProcessCond
	log.Info("Wait for running state of processes.", "processes", waitForRunning)
	for i := range waitForRunning {
		conds = append(conds, pc.ProcessCond{Name: waitForRunning[i], State: pc.ProcessRunning})
	}

	log.Info("Wait for readiness state processes.", "processes", waitForReady)
	for i := range waitForReady {
		conds = append(conds, pc.ProcessCond{Name: waitForReady[i], State: pc.ProcessReady})
	}

	isRunning, err := pcCtx.WaitTill(ctx, 100*time.Millisecond, conds...) //nolint:mnd
	if err != nil || !isRunning {
		return pcCtx, errors.AddContext(err,
			"failed to wait for processes '%q', '%q'", waitForRunning, waitForReady)
	}

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
