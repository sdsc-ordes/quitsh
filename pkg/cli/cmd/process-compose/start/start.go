package processcomposestart

import (
	"context"
	"os"
	"slices"
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
specified by an attribute path (e.g. 'mynamespace.test-dbs') or installable
(e.g. './tools/nix#mynamespace.test-dbs')
in a 'flake.nix' file.`

const timeoutWait = 100 * time.Second
const timeoutWaitInterval = 100 * time.Millisecond

type (
	BasicArgs struct {
		Impl     string
		AttrPath string
		FlakeDir string
	}

	startArgs struct {
		BasicArgs
		socketPathFile string

		waitFor           []string
		waitForReady      []string
		noFailOnCompleted []string
		attach            bool
		timeoutWait       time.Duration
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
			stArgs.AttrPath = args[0]

			_, err := startProcessCompose(
				cl.RootDir(),
				stArgs.FlakeDir,
				stArgs.AttrPath,
				pc.ProcessComposeImpl(stArgs.Impl),
				stArgs.waitFor,
				stArgs.waitForReady,
				stArgs.noFailOnCompleted,
				stArgs.socketPathFile,
				stArgs.timeoutWait,
				stArgs.attach)

			return err
		},
	}

	DefineBasicArgs(startCmd, &stArgs.BasicArgs, defaultFlakeDir)

	startCmd.Flags().
		StringArrayVarP(&stArgs.waitFor,
			"wait-for", "w", nil, "Wait for these processes to be running.")

	startCmd.Flags().
		StringArrayVarP(&stArgs.waitForReady,
			"wait-for-ready", "r", nil, "Wait for these processes to be ready.")

	startCmd.Flags().
		StringArrayVar(&stArgs.noFailOnCompleted,
			"no-fail-on-completed", nil,
			"By default if a '--wait-for' or '--wait-for-ready' completes, the wait condition fails. "+
				"Turn this off for certain processes.")

	startCmd.Flags().
		StringVarP(&stArgs.socketPathFile,
			"socket-path-file", "s", "", "The file where the process-compose socket path is written to.")

	startCmd.Flags().
		BoolVarP(&stArgs.attach,
			"attach", "a", false, "If after start we attach to the process-compose instance.")

	startCmd.Flags().
		DurationVar(&stArgs.timeoutWait,
			"timeout", timeoutWait, "The max. timeout (e.g. `100s`) for waiting on processes.")

	parent.AddCommand(startCmd)
}

func DefineBasicArgs(cmd *cobra.Command, baseArgs *BasicArgs, defaultFlakeDir string) {
	cmd.Flags().
		StringVarP(&baseArgs.FlakeDir,
			"flake-dir", "f", defaultFlakeDir, "The flake directory which contains a 'flake.nix' file.")

	cmd.Flags().
		StringVarP(&baseArgs.Impl,
			"impl", "i", string(pc.ProcessComposeOverServicesFlake),
			"Use `devenv` if the attribute is a `devenv` Nix shell "+
				"or a `services-flake` if it is a `services-flake`-derivation.")
}

// startProcessCompose starts the process-compose services from `flake.nix` in `flakeDir`
// defined in the installable `devenvShellInstallable`.
// You can wait for the processes names to be running with `waitFor`.
//
//nolint:funlen
func startProcessCompose(
	rootDir string,
	flakeDir string,
	attrPath string,
	impl pc.ProcessComposeImpl,
	waitForRunning []string,
	waitForReady []string,
	noFailOnCompleted []string,
	socketPathFile string,
	timeoutWait time.Duration,
	attach bool,
) (
	pcCtx *pc.ProcessComposeCtx,
	err error,
) {
	if strings.Contains(attrPath, "#") {
		pcCtx, err = pc.StartFromInstallable(
			log.Global(),
			rootDir,
			attrPath,
			impl,
			pc.WithMustBeStarted(false),
		)
	} else {
		pcCtx, err = pc.Start(
			log.Global(),
			rootDir,
			flakeDir,
			attrPath,
			impl,
			pc.WithMustBeStarted(false),
		)
	}
	if err != nil {
		return pcCtx, errors.AddContext(err, "could not start process-compose")
	}
	defer func() {
		if err != nil {
			log.Warn("Stopping due to errors.")
			e := pcCtx.Stop()

			log.ErrorE(e, "Could not stop process-compose.")
		}
	}()

	if socketPathFile != "" {
		log.Infof("Write socket path file '%s'.", socketPathFile)
		err = os.WriteFile(socketPathFile, []byte(pcCtx.Socket()), fs.DefaultPermissionsFile)
		if err != nil {
			log.WarnE(err, "Could not write socket path to file '%s'.", socketPathFile)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutWait)
	defer cancel()

	var conds []pc.ProcessCond
	for i := range waitForRunning {
		noFailOnCompleted := slices.Contains(noFailOnCompleted, waitForRunning[i])
		conds = append(conds,
			pc.ProcessCond{
				Name:              waitForRunning[i],
				State:             pc.ProcessRunning,
				NoFailOnCompleted: noFailOnCompleted})
	}

	for i := range waitForReady {
		noFailOnCompleted := slices.Contains(noFailOnCompleted, waitForReady[i])
		conds = append(conds,
			pc.ProcessCond{
				Name:              waitForReady[i],
				State:             pc.ProcessReady,
				NoFailOnCompleted: noFailOnCompleted},
		)
	}

	if len(conds) != 0 {
		log.Info("Wait for processes.",
			"ready", waitForReady,
			"running", waitForRunning,
			"timeout", timeoutWait,
			"interval", timeoutWaitInterval)
	}

	fulfilled, err := pcCtx.WaitTill(ctx, log.Global(), conds...)
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
