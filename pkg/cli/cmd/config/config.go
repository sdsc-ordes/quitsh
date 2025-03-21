package configcmd

import (
	"errors"

	printcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/config/print"
	writecmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/config/write"
	"github.com/sdsc-ordes/quitsh/pkg/config"

	"github.com/spf13/cobra"
)

func AddCmd(parent *cobra.Command, config config.IConfig) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Config sub-commands.",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return errors.New("no subcommand given")
		},
	}

	writecmd.AddCmd(configCmd, config)
	printcmd.AddCmd(configCmd, config)

	parent.AddCommand(configCmd)

	return configCmd
}
