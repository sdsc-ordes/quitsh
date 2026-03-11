package fs

import (
	"errors"
	stderr "errors"
	stdfs "io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/charlievieth/fastwalk"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// FindFiles finds all files in `rootDir`.
// Note: Take care when using [WithPathFilter] or [WithPathFilterPatterns] because
// they will influence how the files are walked.
// This function uses goroutines under to hood.
func FindFiles(
	rootDir string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	return FindPaths(rootDir, append(opts, WithReportOnlyDirs(false))...)
}

// FilterFiles is the same as [FindFiles] but for a given list.
func FilterFiles(
	paths []string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	return FilterPaths(paths, append(opts, WithReportOnlyDirs(false))...)
}

// FindFilesByPatterns finds all files in `rootDir` which match glob patterns
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
			WithPathFilterPatterns(includeGlobPatterns, excludeGlobPatterns, true),
			WithReportOnlyDirs(false))...,
	)
}

// FilterFilesByPatterns is the same as [FindFilesByPatterns] but for a given list.
func FilterFilesByPatterns(
	paths []string,
	includeGlobPatterns []string,
	excludeGlobPatterns []string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	return FilterPaths(paths,
		append(opts,
			WithPathFilterPatterns(includeGlobPatterns, excludeGlobPatterns, true),
			WithReportOnlyDirs(false))...,
	)
}

// FindPaths finds all paths in `rootDir`.
// Note: Take care when using [WithPathFilter] or [WithPathFilterPatterns] because
// they will influence how the files are walked.
// This function uses goroutines under to hood.
func FindPaths(
	rootDir string,
	opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	var query queryOptions
	err = query.Apply(opts)
	if err != nil {
		return nil, 0, err
	}

	// Always use the default walk filter to
	// ignore annoying directories.
	if query.walkDirFilter == nil {
		err = WithWalkDirFilterDefault(true)(&query)
		if err != nil {
			return files, traversedFiles, err
		}
	}

	conf := fastwalk.Config{
		ToSlash: true,
	}

	// TODO: Use a channel and a collecting goroutine here instead of locking etc.
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
			query.walkDirFilter,
			query.pathFilter,
		),
	)

	// Weird workaround, since the Walk might return the last skipdir err
	if stderr.Is(err, fastwalk.SkipDir) {
		err = nil
	}

	traversedFiles = countA.Load()

	return files, traversedFiles, err
}

// FilterPaths is the same as [FindPaths] but for given paths.
func FilterPaths(paths []string, opts ...FindOptions,
) (files []string, traversedFiles int64, err error) {
	var query queryOptions
	err = query.Apply(opts)
	if err != nil {
		return nil, 0, err
	}

	countA := atomic.Int64{}

	walkFunc := createVisitor(
		&countA,
		nil,
		&files,
		query.dirsOnly,
		query.walkDirFilter,
		query.pathFilter,
	)

	for i := range paths {
		fI, e := os.Lstat(paths[i])
		if e != nil {
			return nil, 0, e
		}

		dirE := stdfs.FileInfoToDirEntry(fI)
		e = walkFunc(paths[i], dirE, nil)
		if e != nil && !errors.Is(e, filepath.SkipDir) {
			// Should only return skip, dir, otherwise report.
			return nil, 0, e
		}
	}

	traversedFiles = countA.Load()

	return files, traversedFiles, nil
}

//nolint:gocognit
func createVisitor(
	count *atomic.Int64,
	lock *sync.Mutex,
	paths *[]string,
	dirsOnly bool,
	walkDirFilter PathFilter,
	pathFilter PathFilter,
) stdfs.WalkDirFunc {
	return func(p string, info os.DirEntry, err error) error {
		log.Tracef("Visit path: '%s'", p)

		if err != nil {
			// on permission errors we just skip it
			return filepath.SkipDir
		}

		// Check on directories if we skip this dir and also dont
		// report it.
		if info.IsDir() && walkDirFilter != nil {
			match := walkDirFilter(p, info)
			if !match {
				return filepath.SkipDir
			}
		}

		match := true
		if pathFilter != nil {
			match = pathFilter(p, info)
		}
		trace("Path match: '%s', '%v'", p, match)

		var add bool
		if info.IsDir() {
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
			if lock != nil {
				lock.Lock()
			}
			*paths = append(*paths, p)
			if lock != nil {
				lock.Unlock()
			}
		}

		return nil
	}
}
