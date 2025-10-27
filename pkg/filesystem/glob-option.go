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
		walkDirFilter PathFilter
		pathFilter    PathFilter
		dirsOnly      bool
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

// WithWalkDirFilter sets a custom walk filter
// to determine which directories to skip.
func WithWalkDirFilter(f PathFilter, useAnd bool) FindOptions {
	return func(o *queryOptions) error {
		if o.walkDirFilter != nil {
			old := o.walkDirFilter
			o.walkDirFilter = func(p string, i os.DirEntry) bool {
				if useAnd {
					return old(p, i) && f(p, i)
				} else {
					return old(p, i) || f(p, i)
				}
			}
		} else {
			o.walkDirFilter = f
		}

		return nil
	}
}

// IgnoredDirectoriesDefault returns all by default ignored directories.
// with outputs.
func IgnoredDirectoriesDefault() []string {
	return append(IgnoredDirectoriesDefaultBasic(), OutputDir)
}

// IgnoredDirectoriesDefaultBasic returns all by default ignored directories (basic)
// without output folders.
func IgnoredDirectoriesDefaultBasic() []string {
	return []string{".git", ".direnv", ".devenv", ".venv"}
}

// WithWalkDirFilterDefault sets the default walk directory filter if non it set.
// Note: This is not the same as the `[WithPathFilter]` counter part as
// it determines the discovery of files when walking the tree.
// `useAnd` will logically and this filter to a default one if set.
func WithWalkDirFilterDefault(useAnd bool) FindOptions {
	var def = IgnoredDirectoriesDefault()

	f := func(p string, _ os.DirEntry) bool {
		return !slices.Contains(def, path.Base(p))
	}

	return WithWalkDirFilter(f, useAnd)
}

// WithWalkDirFilterPatterns sets doublestar glob patterns to skip directories when
// walking over the filesystem.
func WithWalkDirFilterPatterns(include []string, exclude []string, useAnd bool) FindOptions {
	f := func(p string, _ os.DirEntry) bool { return MatchByPatterns(p, include, exclude) }

	return WithWalkDirFilter(f, useAnd)
}

// WithPathFilterPatterns sets the doublestar glob patterns path filter
// on discovered paths.
//
// `useAnd` will logically and this filter to a default one if set.
func WithPathFilterPatterns(include []string, exclude []string, useAnd bool) FindOptions {
	f := func(p string, _ os.DirEntry) bool { return MatchByPatterns(p, include, exclude) }

	return WithPathFilter(f, useAnd)
}

// WithReportOnlyDirs sets to only walk directories (no files) which means any
// matching filters are done on directory paths only.
func WithReportOnlyDirs(enable bool) FindOptions {
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
