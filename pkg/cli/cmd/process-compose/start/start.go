package processcomposestart

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	processcompose "github.com/sdsc-ordes/quitsh/pkg/exec/process-compose"
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
	attrPath string
	flakeDir string
	waitFor  []string

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
			"wait-for", "w", nil, "Wait for this processes to be running.")

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
	waitFor []string,
	socketPathFile string,
	attach bool) (
	pcCtx processcompose.ProcessComposeCtx,
	err error,
) {
	if strings.Contains(devenvShellAttrPath, "#") {
		pcCtx, err = processcompose.StartFromInstallable(
			rootDir,
			devenvShellAttrPath,
			false,
		)
	} else {
		pcCtx, err = processcompose.Start(rootDir, flakeDir, devenvShellAttrPath, false)
	}
	if err != nil {
		return pcCtx, errors.AddContext(err, "could not start process-compose")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutContainers)
	defer cancel()

	isRunning, err := pcCtx.WaitTillRunning(ctx, waitFor...)
	if err != nil || !isRunning {
		return pcCtx, errors.AddContext(err, "failed to wait for processes '%q'.", waitFor)
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
