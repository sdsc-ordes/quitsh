package query

import (
	"os"
	"path"
	"slices"

	"github.com/sdsc-ordes/quitsh/pkg/component"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
)

type (
	CompFilter func(_compName, _root string) (_matches bool)

	queryOptions struct {
		configFileName string
		compFilter     CompFilter
		fsOpts         []fs.FindOptions
	}

	Option func(opts *queryOptions) error
)

func newQueryOptions() queryOptions {
	return queryOptions{configFileName: component.ConfigFilename}
}

// Apply applies all options to the config.
func (o *queryOptions) Apply(opts []Option) error {
	for i := range opts {
		if opts[i] == nil {
			continue
		}

		err := opts[i](o)
		if err != nil {
			return err
		}
	}

	return nil
}

// WithFindOptions sets [fs.FindOptions] for the search over the directories.
func WithFindOptions(opts ...fs.FindOptions) Option {
	return func(o *queryOptions) error {
		o.fsOpts = opts

		return nil
	}
}

// WithComponentDirSingle returns a simple filter which only returns
// the component with root directory `compDir`.
func WithComponentDirSingle(compDir string, useAnd bool) Option {
	f := func(_compName, root string) bool {
		return fs.MakeAbsolute(compDir) == root
	}

	return WithCompDirFilter(f, useAnd)
}

// WithCompDirFilter combines a component filter.
func WithCompDirFilter(f CompFilter, useAnd bool) Option {
	return func(o *queryOptions) error {
		if o.compFilter != nil {
			old := o.compFilter
			o.compFilter = func(c, r string) bool {
				if useAnd {
					return old(c, r) && f(c, r)
				} else {
					return old(c, r) || f(c, r)
				}
			}
		} else {
			o.compFilter = f
		}

		return nil
	}
}

// WithCompDirPatternsCombined is the same as WithCompDirPatterns but with exclude syntax `!<pattern>`.
func WithCompDirPatternsCombined(patterns []string, useAnd bool) Option {
	incls, excls := splitIntoIncludeAndExcludes(patterns)

	return WithCompDirPatterns(incls, excls, useAnd)
}

// WithCompDirPatterns add a component filter based on name patterns.
func WithCompDirPatterns(incls []string, excls []string, useAnd bool) Option {
	filt := func(name string, _ string) bool {
		return fs.MatchByPatterns(name, incls, excls)
	}

	return WithCompDirFilter(filt, useAnd)
}

// WithComponentConfigFilename sets the components config filename to be used
// (default is `comp.ConfigFileName`).
func WithComponentConfigFilename(filename string) Option {
	return func(o *queryOptions) error {
		o.configFileName = filename

		return nil
	}
}

// withWalkDirFilterDefault sets the default path filter if non it set.
// `useAnd` will logically and this  to a default one if set.
func withWalkDirFilterDefault(useAnd bool) fs.FindOptions {
	var def = []string{"external"}

	f := func(p string, _ os.DirEntry) bool {
		return !slices.Contains(def, path.Base(p))
	}

	return fs.WithWalkDirFilter(f, useAnd)
}
