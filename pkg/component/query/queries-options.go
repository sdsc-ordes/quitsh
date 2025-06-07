package query

import (
	"path"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type (
	CompFilter func(_compName, _root string) (_matches bool)

	queryOptions struct {
		configFileName string
		compFilter     CompFilter
	}

	Option func(opts *queryOptions) error
)

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

// SingleComponentDirFilter returns a simple filter which only returns
// the component with root directory `compDir`.
func SingleComponentDirFilter(compDir string) Option {
	f := func(_compName, root string) bool {
		return fs.MakeAbsolute(compDir) == root
	}

	return WithCompDirFilter(f, true)
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

// WithCompDirPatterns add a component filter based on name patterns.
func WithCompDirPatterns(incls []string, excls []string, useAnd bool) Option {
	filt := func(name string, _ string) bool {
		include := false
		exclude := false

		for _, pattern := range incls {
			matches, err := doublestar.Match(pattern, name)
			if err != nil {
				log.Warnf("Pattern '%s' is invalid.", pattern)

				return false
			} else if matches {
				include = true

				break
			}
		}

		for _, pattern := range excls {
			matches, err := doublestar.Match(pattern, name)
			if err != nil {
				log.Warnf("Pattern '%s' is invalid.", pattern)

				return false
			} else if matches {
				exclude = true

				break
			}
		}

		return include && !exclude
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

// withPathFilterDefault sets the default path filter if non it set.
// `useAnd` will logically and this  to a default one if set.
func withPathFilterDefault(useAnd bool) fs.FindOptions {
	var def = []string{"external"}

	f := func(p string) bool {
		return !slices.Contains(def, path.Base(p))
	}

	return fs.WithPathFilter(f, useAnd)
}
