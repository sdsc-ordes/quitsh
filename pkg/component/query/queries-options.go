package query

import (
	"path"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
)

type Option func(opts *queryOptions) error

type CompFilter func(_compName, _root string) (_matches bool, _err error)

// WithPathFilter set a custom path filter.
func WithPathFilter(filter fs.DirFilter) Option {
	return func(o *queryOptions) error {
		o.dirFilter = filter

		return nil
	}
}

// WithPathFilterDefault sets the default path filter if non it set.
func WithPathFilterDefault() Option {
	return func(o *queryOptions) error {
		o.dirFilter = func(dir string) bool {
			d := path.Base(dir)

			return d != ".git" && d != ".direnv"
		}

		return nil
	}
}

// ComponentDirFilter returns a simple filter which only returns
// the component with root directory `compDir`.
func ComponentDirFilter(compDir string) CompFilter {
	return func(_compName, root string) (matches bool, err error) {
		if fs.MakeAbsolute(compDir) == root {
			return true, nil
		}

		return
	}
}

// WithFilterAnd sets another component filter `f` in an `g && f` combination.
func WithFilterAnd(f CompFilter) Option {
	return func(o *queryOptions) error {
		if o.compFilter == nil {
			o.compFilter = f
		} else {
			// Do a mixing.
			old := o.compFilter
			o.compFilter = func(compName, root string) (matches bool, err error) {
				matches, err = old(compName, root)
				if !matches || err != nil {
					return
				}

				return f(compName, root)
			}
		}

		return nil
	}
}

// WithComponentConfigFilename sets the components config filename to be used
// (default is `comp.ConfigFileName`).
func WithComponentConfigFilename(filename string) Option {
	return func(o *queryOptions) error {
		o.configFileName = filename

		return nil
	}
}
