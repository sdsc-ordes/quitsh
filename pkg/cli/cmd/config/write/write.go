package writecmd

import (
	"github.com/sdsc-ordes/quitsh/pkg/config"

	"github.com/spf13/cobra"
)

func AddCmd(root *cobra.Command, config config.IConfig) {
	var output string

	ciCmd := &cobra.Command{
		Use:   "write",
		Short: "Write the current config file to disk",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return outputConfig(output, config)
		},
	}

	ciCmd.Flags().
		StringVarP(&output, "output", "o", "quitsh.yaml", "The output path of the config.")
	_ = ciCmd.MarkFlagRequired("output")

	root.AddCommand(ciCmd)
}

func outputConfig(path string, conf config.IConfig) error {
	return config.SaveInterfaceToFile(path, conf)
}
