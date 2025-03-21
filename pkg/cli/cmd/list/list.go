package listcmd

import (
	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/spf13/cobra"
)

type listArgs struct {
	compArgs general.ComponentArgs
}

const longDesc = `
List all components found in the current working directory.
`

func AddCmd(cl cli.ICLI, parent *cobra.Command) {
	var args listArgs

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List components",
		Long:  longDesc,
		PreRunE: func(_cmd *cobra.Command, _args []string) error {
			return nil
		},
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return listComponents(cl, &args)
		},
	}

	listCmd.Flags().
		StringArrayVarP(&args.compArgs.ComponentPatterns,
			"components", "c", []string{"*"}, "Components matched by these patterns are listed.")

	parent.AddCommand(listCmd)
}

func listComponents(cl cli.ICLI, c *listArgs) error {
	comps, _, _, err := cl.FindComponents(&c.compArgs)

	if err != nil {
		return err
	}

	for i := range comps {
		log.Info(
			"Component:",
			"root",
			comps[i].Root(),
			"name",
			comps[i].Config().Name,
			"version",
			comps[i].Config().Version.String(),
		)
	}

	return nil
}
