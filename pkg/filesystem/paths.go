package fs

import (
	"os"
	"path"
	"path/filepath"

	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// These constants define general directories used throughout
// the repository.
const (
	OutputDir = ".output"

	// Relative paths to `OutputDir`.
	OutBuildDir      = "build"
	OutBuildBinDir   = "build/bin"
	OutBuildShareDir = "build/share"
	OutBuildDocsDir  = "build/docs"

	OutPackageDir = "package"

	OutCoverageDir     = "coverage"
	OutCoverageDataDir = "coverage/data"
	OutCoverageBinDir  = "coverage/bin"

	OutRunDir = "run"

	OutImageDir = OutPackageDir

	OutCIDir = "ci"

	DocsDir   = "docs"
	ImagesDir = "images"
)

// Exists checks if a path exists. Follows symlinks.
func Exists(path string) (exists bool) {
	exists, _ = ExistsE(path)

	return
}

// ExistsE checks if a path exists and returns any error associated with
// `os.Stat`. It follows symlinks.
func ExistsE(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ExistsL checks if a path exists. It does not follow symlink.
func ExistsL(path string) (exists bool) {
	exists, _ = ExistsLE(path)

	return
}

// ExistsLE checks if a path exists and returns any error
// associated with `os.Stat`. It does not follow symlinks.
func ExistsLE(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	return true, nil
}

// FindRelPathInParents finds the existing relative path `p` (e.g `a/b/c/test`)
// (if not relative the absolute is returned)
// starting from `start` (can be relative)
// against all parents until (and including) the `root`
// is discovered or no more parents exist.
func FindRelPathInParents(start string, p string, root string) (found string) {
	if path.IsAbs(p) {
		if Exists(p) {
			return p
		} else {
			return ""
		}
	}

	if root != "" {
		root = MakeAbsolute(root)
	}
	dir := MakeAbsolute(start)

	for {
		file := path.Join(dir, p)
		if Exists(file) {
			found = file

			break
		} else if root != "" && dir == root {
			break
		}

		n := path.Dir(dir) // `n` is cleaned.
		if n == dir {
			break
		}
		dir = n
	}

	return found
}

// MakeAbsoluteTo makes a path absolute to the `base` directory.
func MakeAbsoluteTo(base string, p string) string {
	debug.Assert(path.IsAbs(base), "Base path is not absolute.")

	if !path.IsAbs(p) {
		p = path.Join(base, p)
	}

	return p
}

// MakeAbsolute makes path `p` absolute to the current working directory.
func MakeAbsolute(p string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.PanicE(err, "Could not evaluate cwd.")
	}

	return MakeAbsoluteTo(cwd, p)
}

// MakeAllAbsolute makes all paths absolute to the current working directory.
// This function works inplace!
func MakeAllAbsolute(p ...string) []string {
	cwd, err := os.Getwd()
	if err != nil {
		log.PanicE(err, "Could not evaluate cwd.")
	}

	for i := range p {
		p[i] = MakeAbsoluteTo(cwd, p[i])
	}

	return p
}

// MakeAllAbsoluteTo makes all paths absolute to the `base` directory.
// This function works inplace!
func MakeAllAbsoluteTo(base string, p ...string) []string {
	for i := range p {
		p[i] = MakeAbsoluteTo(base, p[i])
	}

	return p
}

// MakeRelativeTo makes a `path` relative to `base`.
func MakeRelativeTo(base string, path string) (s string, e error) {
	s, e = filepath.Rel(base, path)
	s = filepath.ToSlash(s)

	return
}
