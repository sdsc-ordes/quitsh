package fs

import (
	"errors"
	stdfs "io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/charlievieth/fastwalk"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// Find all files in `rootDir`.
// Note: Take care when using [WithPathFilter] or [WithGlobPatterns] because
// they will influence how the files are walked.
// This function uses goroutines under to hood.
func FindFiles(
	rootDir string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	return FindPaths(rootDir, append(opts, WithWalkDirsOnly(false))...)
}

// Find all files in `rootDir` which match glob patterns
// `includePatterns` and not `excludePatterns` (they match the full path).
// This function uses goroutines under to hood.
func FindFilesByPatterns(
	rootDir string,
	includeGlobPatterns []string,
	excludeGlobPatterns []string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	return FindPaths(rootDir,
		append(opts,
			WithGlobFilePatterns(includeGlobPatterns, excludeGlobPatterns, true),
			WithWalkDirsOnly(false))...,
	)
}

// Find all paths in `rootDir`.
// Note: Take care when using [WithPathFilter] or [WithGlobPatterns] because
// they will influence how the files are walked.
// This function uses goroutines under to hood.
func FindPaths(
	rootDir string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	var query queryOptions
	for i := range opts {
		if opts[i] == nil {
			continue
		}
		err = opts[i](&query)
		if err != nil {
			return nil, 0, err
		}
	}

	// Always use the default path filter.
	if query.pathFilter == nil {
		err = WithPathFilterDefault(true)(&query)
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
			query.dirsOnly,
			query.pathFilter,
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
	dirsOnly bool,
	pathFilter PathFilter,
) stdfs.WalkDirFunc {
	return func(p string, info os.DirEntry, err error) error {
		log.Tracef("Visit path: '%s'", p)

		if err != nil {
			// on permission errors we just skip it
			return filepath.SkipDir
		}

		match := true
		if pathFilter != nil {
			match = pathFilter(p, info)
		}
		trace("Path match: '%s', '%v'", p, match)

		var add bool
		if info.IsDir() {
			if !match {
				// Directory ignored.
				return filepath.SkipDir
			}

			if dirsOnly {
				count.Add(1)
				add = true
			}
			trace("Dir: '%s'", p)
		} else {
			if !dirsOnly {
				count.Add(1)
				add = true
			}
			trace("File: '%s'", p)
		}

		// Add the path if it matches.
		if add && match {
			lock.Lock()
			*files = append(*files, p)
			lock.Unlock()
		}

		return nil
	}
}
