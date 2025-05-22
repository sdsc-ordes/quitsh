package general

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/query"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// Some component args for the CLI commands.
type ComponentArgs struct {
	// Exclusive arguments:
	// Glob patterns to components.
	// Negation works with `!...`
	ComponentPatterns []string

	// or a destinct component directory.
	ComponentDir string
}

// Find dispatches to the query function to find all components and
// returns them.
func FindComponents(
	args *ComponentArgs,
	rootDir string,
	outDirBase string,
	defaultCompPatterns []string,
	transformConfig component.ConfigAdjuster,
) (comps []*component.Component, all []*component.Component, err error) {
	compCreator := component.NewComponentCreator(outDirBase, transformConfig)

	switch {
	case len(args.ComponentPatterns) != 0:
		comps, all, err = query.FindByPatterns(
			rootDir,
			append(args.ComponentPatterns, defaultCompPatterns...),
			1,
			compCreator,
		)
	case args.ComponentDir != "":
		compDir := fs.MakeAbsolute(args.ComponentDir)

		comps, all, err = query.FindByPatterns(
			rootDir,
			append([]string{"*"}, defaultCompPatterns...),
			1,
			compCreator,
			query.WithFilterAnd(query.ComponentDirFilter(compDir)),
		)

		debug.Assert(len(comps) == 1, "should have length 1")

	default:
		return nil, nil, errors.New("you need to specify at least components patterns " +
			"or a component directory")
	}

	log.Info("Components count:", "all", len(all), "selected", len(comps))

	return
}
