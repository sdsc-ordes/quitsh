package processcomposestop

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	processcompose "github.com/sdsc-ordes/quitsh/pkg/exec/process-compose"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/cobra"
)

const longDesc = `Stop a process-compose definition from a 'devenv.sh' Nix shell
specified by an attribute path (e.g. 'mynamespace.shells.test-dbs') or installable
(e.g. './tools/nix#mynamespace.shells.test-dbs')
in a 'flake.nix' file.`

type startArgs struct {
	attrPath string
	flakeDir string
}

func AddCmd(cl cli.ICLI, parent *cobra.Command, defaultFlakeDir string) {
	var stArgs startArgs

	startCmd := &cobra.Command{
		Use:     "stop [devenv-attr-path or devenv-installable]",
		Short:   "Stop a process-compose definition from a 'devenv.sh' Nix shell.",
		Long:    longDesc,
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			stArgs.attrPath = args[0]

			_, err := StopService(
				cl.RootDir(),
				stArgs.flakeDir,
				stArgs.attrPath)

			return err
		},
	}

	startCmd.Flags().
		StringVarP(&stArgs.flakeDir,
			"flake-dir", "f", defaultFlakeDir, "The flake directory which contains a 'flake.nix' file.")

	parent.AddCommand(startCmd)
}

// StartServices starts the process-compose services from `flake.nix` in `flakeDir`
// defined in the installable `devenvShellInstallable`.
// You can wait for the processes names to be running with `waitFor`.
func StopService(
	rootDir string,
	flakeDir string,
	devenvShellAttrPath string) (
	pcCtx processcompose.ProcessComposeCtx,
	err error,
) {
	if strings.Contains(devenvShellAttrPath, "#") {
		pcCtx, err = processcompose.StartFromInstallable(
			rootDir,
			devenvShellAttrPath,
			true,
		)
	} else {
		pcCtx, err = processcompose.Start(rootDir, flakeDir, devenvShellAttrPath, true)
	}

	if err != nil {
		return pcCtx, errors.AddContext(err, "could not get process-compose instance")
	}

	err = pcCtx.Stop()
	if err != nil {
		return pcCtx, errors.AddContext(err, "error occurred in stopping process-compose instance")
	}

	log.Infof("Stopped process-compose instance '%s'.", pcCtx.Socket())

	return pcCtx, nil
}
