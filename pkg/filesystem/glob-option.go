package fs

import (
	"os"
	"path"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

const traceEnabled = build.DebugEnabled

func trace(msg string, args ...any) {
	if traceEnabled {
		log.Tracef(msg, args...)
	}
}

// PathFilter is the filter which returns true for
// directories to recurse into when querying.
type (
	PathFilter func(path string, info os.DirEntry) bool

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
			o.pathFilter = func(p string, i os.DirEntry) bool {
				if useAnd {
					return old(p, i) && f(p, i)
				} else {
					return old(p, i) || f(p, i)
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

	f := func(p string, _ os.DirEntry) bool {
		return !slices.Contains(def, path.Base(p))
	}

	return WithPathFilter(f, useAnd)
}

// WithGlobPatterns sets the doublestar glob patterns path filter.
// Note: If you want to filter files use [WithGlobFilePatterns] or
//
//	[WithGlobDirPatterns] for directories. This patterns are used when
//	walking through the filesystem, any ignore on a directory will result in not finding
//	other files.
//
// `useAnd` will logically and this filter to a default one if set.
func WithGlobPatterns(include []string, exclude []string, useAnd bool) FindOptions {
	f := func(p string, _ os.DirEntry) bool { return MatchByPatterns(p, include, exclude) }

	return WithPathFilter(f, useAnd)
}

// WithGlobFilePatterns sets the doublestar glob patterns for files, (all directory paths are not touched).
// `useAnd` will logically and this filter to a default one if set.
func WithGlobFilePatterns(include []string, exclude []string, useAnd bool) FindOptions {
	f := func(p string, i os.DirEntry) bool { return i.IsDir() || MatchByPatterns(p, include, exclude) }

	return WithPathFilter(f, useAnd)
}

// WithGlobDirPattern sets the doublestar glob patterns for directory, (all directory paths are not touched).
// `useAnd` will logically and this filter to a default one if set.
func WithGlobDirPatterns(include []string, exclude []string, useAnd bool) FindOptions {
	f := func(p string, i os.DirEntry) bool { return !i.IsDir() || MatchByPatterns(p, include, exclude) }

	return WithPathFilter(f, useAnd)
}

// WithWalkDirsOnly sets to only walk directories (no files) which means any
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
// If `includeGlobs` is empty, it acts as `*` include all.
func MatchByPatterns(s string, includeGlobs, excludeGlobs []string) bool {
	include := len(includeGlobs) == 0 // include all
	exclude := false

	if !include {
		for _, pattern := range includeGlobs {
			matches, err := doublestar.Match(pattern, s)

			if err != nil {
				log.Warnf("Include pattern '%s' is invalid.", pattern)

				return false
			} else if matches {
				trace("Match include: '%s'", pattern)
				include = true

				break
			}
		}
	}

	for _, pattern := range excludeGlobs {
		matches, err := doublestar.Match(pattern, s)

		if err != nil {
			log.Warnf("Exclude pattern '%s' is invalid.", pattern)

			return false
		} else if matches {
			trace("Match exclude: '%s'", pattern)
			exclude = true

			break
		}
	}

	return include && !exclude
}
