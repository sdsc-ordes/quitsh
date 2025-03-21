package fs

import (
	"errors"
	stdfs "io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/charlievieth/fastwalk"
)

// Find all files in `rootDir` which match glob patterns
// `includePatterns` and not `excludePatterns` (they match the full path).
// This function uses goroutines under to hood.
func FindFiles(
	rootDir string,
	includeGlobPatterns []string,
	excludeGlobPatterns []string,
	opts ...Option,
) (files []string, traversedFiles int64, err error) {
	var query queryOptions
	for i := range opts {
		err = opts[i](&query)
		if err != nil {
			return files, traversedFiles, err
		}
	}

	// Always use the default dir filter.
	if query.dirFilter == nil {
		err = WithPathFilterDefault()(&query)
		if err != nil {
			return files, traversedFiles, err
		}
	}

	conf := fastwalk.Config{
		ToSlash: true,
	}

	lock := sync.Mutex{}
	countA := atomic.Int64{}

	err = fastwalk.Walk(
		&conf,
		rootDir,
		createVisitor(
			&countA,
			&lock,
			&files,
			query.dirFilter,
			includeGlobPatterns,
			excludeGlobPatterns,
			query.fileNameOnly,
		),
	)

	// Weird workaround, since the Walk might return the last skipdir err
	if errors.Is(err, fastwalk.SkipDir) {
		err = nil
	}

	traversedFiles = countA.Load()

	return files, traversedFiles, err
}

func createVisitor(
	count *atomic.Int64,
	lock *sync.Mutex,
	files *[]string,
	dirFilter DirFilter,
	includeGlobPatterns []string,
	excludeGlobPatterns []string,
	fileNameOnly bool,
) stdfs.WalkDirFunc {
	return func(p string, info os.DirEntry, err error) error {
		count.Add(1)

		if err != nil {
			// on permission errors we just skip it
			return filepath.SkipDir
		}

		f := p
		if fileNameOnly {
			f = path.Base(p)
		}

		if info.IsDir() {
			if dirFilter != nil && !dirFilter(p) {
				return filepath.SkipDir
			}
		} else if MatchByPatterns(f, includeGlobPatterns, excludeGlobPatterns) {
			lock.Lock()
			*files = append(*files, p)
			lock.Unlock()
		}

		return nil
	}
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
