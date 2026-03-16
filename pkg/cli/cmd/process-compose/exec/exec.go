package processcomposeexec

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	pc "github.com/sdsc-ordes/quitsh/pkg/exec/process-compose"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/cobra"
)

type startArgs struct {
	args []string

	attrPath string
	flakeDir string
}

func AddCmd(cl cli.ICLI, parent *cobra.Command, defaultFlakeDir string) {
	var stArgs startArgs

	startCmd := &cobra.Command{
		Use:     "exec [devenv-attr-path or devenv-installable] [args-to-proc-compose]",
		Short:   "Exec commands on process-compose on the correct instance.",
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			stArgs.attrPath = args[0]
			if len(args) > 1 {
				stArgs.args = args[1:]
			}

			_, err := RunExec(
				log.Global(),
				cl.RootDir(),
				stArgs.flakeDir,
				stArgs.attrPath,
				stArgs.args)

			return err
		},
	}

	startCmd.Flags().
		StringVarP(&stArgs.flakeDir,
			"flake-dir", "f", defaultFlakeDir, "The flake directory which contains a 'flake.nix' file.")

	parent.AddCommand(startCmd)
}

// RunExec runs process-compose commands on the instance
// defined in `devenvShellAttrPath`.
func RunExec(
	log log.ILog,
	rootDir string,
	flakeDir string,
	devenvShellAttrPath string,
	args []string,
) (
	pcCtx *pc.ProcessComposeCtx,
	err error,
) {
	if strings.Contains(devenvShellAttrPath, "#") {
		pcCtx, err = pc.StartFromInstallable(
			log,
			rootDir,
			devenvShellAttrPath,
			pc.WithMustBeStarted(true))
	} else {
		pcCtx, err = pc.Start(
			log,
			rootDir,
			flakeDir,
			devenvShellAttrPath,
			pc.WithMustBeStarted(true))
	}

	if err != nil {
		return pcCtx, errors.AddContext(err, "could not start process-compose")
	}

	return pcCtx, pcCtx.Check(args...)
}
