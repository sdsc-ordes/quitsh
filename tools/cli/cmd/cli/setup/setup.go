package setup

import (
	setup "quitsh-cli/pkg/setup"

	"github.com/spf13/cobra"
)

func AddCmd(root *cobra.Command) {
	setupCmd := &cobra.Command{
		Use:     "setup-development",
		Aliases: []string{"setup"},
		Short:   "Setup local development.",
		Long:    "Setup the repository for local development.",
		PreRunE: func(_cmd *cobra.Command, _args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, _args []string) error {
			return Execute(cmd)
		},
	}

	root.AddCommand(setupCmd)
}

func Execute(
	_command *cobra.Command,
) error {
	return setup.Setup()
}
