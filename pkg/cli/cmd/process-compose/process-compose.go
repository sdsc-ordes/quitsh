package processcompose

import (
	"errors"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	processcomposestart "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/process-compose/start"
	processcomposestop "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/process-compose/stop"

	"github.com/spf13/cobra"
)

func AddCmd(cl cli.ICLI, parent *cobra.Command, flakeDirDefault string) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "process-compose",
		Short: "Start/stop process-compose services defined in Nix 'devenv.sh' shells.",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return errors.New("no subcommand given")
		},
	}

	processcomposestart.AddCmd(cl, configCmd, flakeDirDefault)
	processcomposestop.AddCmd(cl, configCmd, flakeDirDefault)

	parent.AddCommand(configCmd)

	return configCmd
}
