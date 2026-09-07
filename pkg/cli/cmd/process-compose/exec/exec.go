package processcomposeexec

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	processcomposestart "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/process-compose/start"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	pc "github.com/sdsc-ordes/quitsh/pkg/exec/process-compose"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/cobra"
)

type startArgs struct {
	processcomposestart.BasicArgs
	args []string
}

func AddCmd(cl cli.ICLI, parent *cobra.Command, defaultFlakeDir string) {
	var stArgs startArgs

	startCmd := &cobra.Command{
		Use: "exec [attr-path or installable] [args-to-proc-compose]",
		Short: "Exec process-compose definition from a 'devenv.sh' Nix shell or " +
			"a 'process-compose-flake' derivation.",
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			stArgs.AttrPath = args[0]
			if len(args) > 1 {
				stArgs.args = args[1:]
			}

			_, err := RunExec(
				log.Global(),
				cl.RootDir(),
				stArgs.FlakeDir,
				stArgs.AttrPath,
				pc.ProcessComposeImpl(stArgs.Impl),
				stArgs.args,
			)

			return err
		},
	}

	processcomposestart.DefineBasicArgs(startCmd, &stArgs.BasicArgs, defaultFlakeDir)

	parent.AddCommand(startCmd)
}

// RunExec runs process-compose commands on the instance
// defined in `attrPath`.
func RunExec(
	log log.ILog,
	rootDir string,
	flakeDir string,
	attrPath string,
	impl pc.ProcessComposeImpl,
	args []string,
) (
	pcCtx *pc.ProcessComposeCtx,
	err error,
) {
	if strings.Contains(attrPath, "#") {
		pcCtx, err = pc.StartFromInstallable(
			log,
			rootDir,
			attrPath,
			impl,
			pc.WithMustBeStarted(true))
	} else {
		pcCtx, err = pc.Start(
			log,
			rootDir,
			flakeDir,
			attrPath,
			impl,
			pc.WithMustBeStarted(true))
	}

	if err != nil {
		return pcCtx, errors.AddContext(err, "could not start process-compose")
	}

	return pcCtx, pcCtx.Check(args...)
}
