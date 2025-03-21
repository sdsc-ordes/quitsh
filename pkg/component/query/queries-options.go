package query

import (
	"path"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
)

type Option func(opts *queryOptions) error

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

// WithComponentConfigFilename sets the components config filename to be used
// (default is `comp.ConfigFileName`).
func WithComponentConfigFilename(filename string) Option {
	return func(o *queryOptions) error {
		o.configFileName = filename

		return nil
	}
}
