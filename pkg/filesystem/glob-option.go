package fs

import (
	"path"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
)

// PathFilter is the filter which returns true for
// directories to recurse into when querying.
type (
	PathFilter func(path string) bool

	// Options to pass to various query function.
	FindOptions func(opts *queryOptions) error

	queryOptions struct {
		pathFilter PathFilter
		dirsOnly   bool
	}
)

// WithPathFilter set a custom path filter.
func WithPathFilter(f PathFilter, useAnd bool) FindOptions {
	return func(o *queryOptions) error {
		if o.pathFilter != nil {
			old := o.pathFilter
			o.pathFilter = func(p string) bool {
				if useAnd {
					return old(p) && f(p)
				} else {
					return old(p) || f(p)
				}
			}
		} else {
			o.pathFilter = f
		}

		return nil
	}
}

// DefaultIgnoredDirectories returns all by default ignored directories.
func IgnoredDirectoriesDefault() []string {
	return []string{".git", ".direnv", ".devenv"}
}

// WithPathFilterDefault sets the default path filter if non it set.
// `useAnd` will logically and this filter to a default one if set.
func WithPathFilterDefault(useAnd bool) FindOptions {
	var def = IgnoredDirectoriesDefault()

	f := func(p string) bool {
		return !slices.Contains(def, path.Base(p))
	}

	return WithPathFilter(f, useAnd)
}

// WithGlobPatterns sets the glob patterns path filter.
// `useAnd` will logically and this filter to a default one if set.
func WithGlobPatterns(include []string, exclude []string, useAnd bool) FindOptions {
	f := func(p string) bool { return MatchByPatterns(p, include, exclude) }

	return WithPathFilter(f, useAnd)
}

// WithWalkDirsOnly sets to only walk directories which means any
// matching filters are done on directory paths only.
func WithWalkDirsOnly(enable bool) FindOptions {
	return func(o *queryOptions) error {
		o.dirsOnly = enable

		return nil
	}
}

// Apply applies all options to the config.
func (o *queryOptions) Apply(opts []FindOptions) error {
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

// Match a string `name` by some include and exclude glob patterns (doublestar allowed).
// All errors of `doublestar.ErrBadPattern` will be ignored for performance reason.
func MatchByPatterns(s string, includeGlobs, excludeGlobs []string) (matches bool) {
	include := false
	exclude := false
	for _, pattern := range includeGlobs {
		matches, _ = doublestar.Match(pattern, s)
		if matches {
			include = true

			break
		}
	}

	for _, pattern := range excludeGlobs {
		matches, _ = doublestar.Match(pattern, s)
		if matches {
			exclude = true

			break
		}
	}

	return include && !exclude
}
