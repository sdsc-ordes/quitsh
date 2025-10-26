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

	OutImageDir = OutPackageDir

	DocsDir   = "docs"
	ImagesDir = "images"

	CIDir = "ci"
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
