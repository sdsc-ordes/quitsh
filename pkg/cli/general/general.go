package general

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/query"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
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
