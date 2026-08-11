package nix

import (
	"errors"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/cmd/nix/cache"
	fixhash "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/nix/fix-hash"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"

	"github.com/spf13/cobra"
)

// AddCmd adds the `nix` subcommands to `rootCmd`.
func AddCmd(cl cli.ICLI, rootCmd *cobra.Command, nixSetts *config.NixSettings) {
	nixCmd := &cobra.Command{
		Use:   "nix",
		Short: "Helper commands for Nix. ",
		RunE: func(cmd *cobra.Command, _args []string) error {
			_ = cmd.Help()

			return errors.New("no command given")
		},
	}

	fixhash.AddCmd(cl, nixCmd, nixSetts)

	cache.AddDownloadCmd(cl, nixCmd, nixSetts)
	cache.AddUploadCmd(cl, nixCmd, nixSetts)

	rootCmd.AddCommand(nixCmd)
}
