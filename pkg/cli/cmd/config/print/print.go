package printcmd

import (
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/spf13/cobra"
)

func AddCmd(parent *cobra.Command, config config.IConfig) {
	configCmd := &cobra.Command{
		Use:   "print",
		Short: "Print the global config.",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return printConfig(config)
		},
	}

	parent.AddCommand(configCmd)
}

func printConfig(config config.IConfig) error {
	log.Info("Config", "config", config)

	return nil
}
