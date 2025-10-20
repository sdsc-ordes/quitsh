package cleancmd

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/spf13/cobra"
)

type cleanArgs struct {
	compArgs   general.ComponentArgs
	gitIgnored bool
	force      bool
}

const longDescClean = `
    Clean outputs of components.
`

func AddCmd(cli cli.ICLI) {
	var args cleanArgs

	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean components.",
		Long:  longDescClean,
		PreRunE: func(_cmd *cobra.Command, _args []string) error {
			return nil
		},
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return execute(cli, &args)
		},
	}

	cleanCmd.Flags().
		StringArrayVarP(&args.compArgs.ComponentPatterns,
			"components", "c", nil, "Components matched by these patterns are built.")
	cleanCmd.Flags().
		StringVar(&args.compArgs.ComponentDir,
			"component-dir", "", "Directory pointing to a component to build, instead of giving them by patterns.")
	cleanCmd.MarkFlagsMutuallyExclusive("components", "component-dir")

	cleanCmd.Flags().
		BoolVarP(&args.gitIgnored,
			"git-ignored", "X", false, "Clean Git ignored files in component dir.")

	cleanCmd.Flags().
		BoolVarP(&args.force,
			"force", "f", false, "Instead of doing a dry-run really clean it.")

	cli.RootCmd().AddCommand(cleanCmd)
}

func execute(cli cli.ICLI, c *cleanArgs) error {
	comps, _, _, err := cli.FindComponents(&c.compArgs)
	if err != nil {
		return err
	}

	for i := range comps {
		comp := comps[i]

		if c.gitIgnored {
			if !c.force {
				log.Info("Dry Run: Would clean with `git clean -X`.", "cwd", comp.Root())

				continue
			}

			log.Info("Cleaning with `git clean -fX`.", "cwd", comp.Root())
			gitx := git.NewCtx(comp.Root())
			err = gitx.Check("clean", "-fX")

			if err != nil {
				log.ErrorE(err, "Git clean failed.")

				return err
			}
		} else {
			outDir := comp.OutDir()
			if !c.force {
				log.Info("Dry Run: Would clean output directory", "path", outDir)

				continue
			}

			log.Info("Removing output directory.", "path", outDir)
			os.RemoveAll(outDir)
		}
	}

	return nil
}
