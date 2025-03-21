package fs

import (
	"path"
	"slices"
)

// DirFilter is the filter which returns true for
// directories to recurse into when querying.
type DirFilter func(path string) bool

// Options to pass to various query function.
type Option func(opts *queryOptions) error

// WithPathFilter set a custom path filter.
func WithPathFilter(filter DirFilter) Option {
	return func(o *queryOptions) error {
		o.dirFilter = filter

		return nil
	}
}

// DefaultIgnoredDirectories returns all by default ignored directories.
func IgnoredDirectoriesDefault() []string {
	return []string{".git", ".direnv"}
}

// WithPathFilterDefault sets the default path filter if non it set.
func WithPathFilterDefault() Option {
	var def = IgnoredDirectoriesDefault()

	return func(o *queryOptions) error {
		o.dirFilter = func(dir string) bool {
			d := path.Base(dir)

			return !slices.Contains(def, d)
		}

		return nil
	}
}

// WithMatchFileNameOnly sets to only match on the filename instead of the full path.
func WithMatchFileNameOnly() Option {
	return func(o *queryOptions) error {
		o.fileNameOnly = true

		return nil
	}
}

type queryOptions struct {
	dirFilter    DirFilter
	fileNameOnly bool
}
