package printcmd

import (
	"github.com/goccy/go-yaml"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/spf13/cobra"
)

func AddCmd(parent *cobra.Command, config config.IConfig) {
	configCmd := &cobra.Command{
		Use:   "print",
		Short: "Print the global config.",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return PrintConfig(config, false)
		},
	}

	parent.AddCommand(configCmd)
}

func PrintConfig(config config.IConfig, asDebug bool) error {
	if asDebug && !log.IsDebug() {
		return nil
	}

	buf, err := yaml.MarshalWithOptions(config, yaml.Indent(2)) //nolint:mnd
	if err != nil {
		return err
	}

	if asDebug {
		log.Debug("Config", "config", string(buf))
	} else {
		log.Info("Config", "config", string(buf))
	}

	return nil
}
