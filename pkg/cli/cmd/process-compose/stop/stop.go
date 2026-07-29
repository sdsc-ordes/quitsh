package processcomposestop

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	processcomposestart "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/process-compose/start"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	pc "github.com/sdsc-ordes/quitsh/pkg/exec/process-compose"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/cobra"
)

const longDesc = `Stop a process-compose definition from a 'devenv.sh' Nix shell
specified by an attribute path (e.g. 'mynamespace.shells.test-dbs') or installable
(e.g. './tools/nix#mynamespace.shells.test-dbs')
in a 'flake.nix' file.`

type startArgs struct {
	processcomposestart.BasicArgs
}

func AddCmd(cl cli.ICLI, parent *cobra.Command, defaultFlakeDir string) {
	var stArgs startArgs

	stopCmd := &cobra.Command{
		Use:     "stop [attr-path to 'devenv' NixShell or 'services-flake' derivation]",
		Short:   "Stop a process-compose definition from a 'devenv.sh' Nix shell.",
		Long:    longDesc,
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			stArgs.AttrPath = args[0]

			_, err := StopService(
				cl.RootDir(),
				stArgs.FlakeDir,
				stArgs.AttrPath,
				pc.ProcessComposeImpl(stArgs.Impl),
			)

			return err
		},
	}

	processcomposestart.DefineBasicArgs(stopCmd, &stArgs.BasicArgs, defaultFlakeDir)

	parent.AddCommand(stopCmd)
}

// StopService stops the process-compose services from `flake.nix` in `flakeDir`
// defined in the attribute `attrPath`.
func StopService(
	rootDir string,
	flakeDir string,
	attrPath string,
	impl pc.ProcessComposeImpl,
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
			pc.WithMustBeStarted(true),
		)
	} else {
		pcCtx, err = pc.Start(
			log.Global(),
			rootDir,
			flakeDir,
			attrPath,
			impl,
			pc.WithMustBeStarted(true))
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
