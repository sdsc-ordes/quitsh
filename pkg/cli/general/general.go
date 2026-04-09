package general

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/query"
	"github.com/sdsc-ordes/quitsh/pkg/dag"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/cobra"
)

// ComponentArgs are arguments for the CLI commands.
type ComponentArgs struct {
	// Exclusive arguments:
	// Glob patterns to components.
	// Negation works with `!...`
	ComponentPatterns []string

	// or a destinct component directory.
	ComponentDir string
}

// AddFlagsComponentArgs adds the flags to command `cmd`
// for an instance of [ComponentArgs].
func AddFlagsComponentArgs(cmd *cobra.Command, compArgs *ComponentArgs) {
	cmd.Flags().
		StringArrayVarP(&compArgs.ComponentPatterns,
			"components", "c", nil, "Components matched by these patterns are built.")
	cmd.Flags().
		StringVar(&compArgs.ComponentDir,
			"component-dir", "", "Directory pointing to a component to build, instead of giving them by patterns.")

	cmd.MarkFlagsMutuallyExclusive("components", "component-dir")
	cmd.MarkFlagsOneRequired("components", "component-dir")
}

// AddFlagsExecArgs adds all `execArgs` arguments to the command.
func AddFlagsExecArgs(cmd *cobra.Command, execArgs *dag.ExecArgs) {
	cmd.Flags().StringArrayVar(&execArgs.Tags, "tag", execArgs.Tags,
		"The executable tags which will get matched against the "+
			"`include.tagExpr` on a step to include/exclude steps.")
}

// FindComponents dispatches to the query function to find all components and
// returns them.
func FindComponents(
	args *ComponentArgs,
	rootDir string,
	outDirBase string,
	transformConfig component.ConfigAdjuster,
	opts ...query.Option,
) (comps []*component.Component, all []*component.Component, err error) {
	compCreator := component.NewComponentCreator(outDirBase, transformConfig)

	switch {
	case len(args.ComponentPatterns) != 0:
		comps, all, err = query.FindByPatterns(
			rootDir,
			args.ComponentPatterns,
			1,
			compCreator,
			opts...,
		)
	case args.ComponentDir != "":
		compDir := fs.MakeAbsolute(args.ComponentDir)

		comps, all, err = query.FindByPatterns(
			rootDir,
			[]string{"*"},
			1,
			compCreator,
			append(opts, query.WithComponentDirSingle(compDir, true))...,
		)

	default:
		return nil, nil, errors.New("you need to specify at least components patterns " +
			"or a component directory")
	}

	log.Info("Components count:", "all", len(all), "selected", len(comps))

	return
}
