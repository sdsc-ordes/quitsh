package fs

import (
	"os"

	"github.com/otiai10/copy"
)

// CopyFileOrDir copies a directory or file from `src` to `dest`.
// If `existsOk` is `true` it will replace `dest` if it exists.
// Copies also symlinks.
func CopyFileOrDir(src string, dest string, existsOk bool) (err error) {
	if !existsOk && ExistsL(dest) {
		return os.ErrExist
	}

	return copy.Copy(src, dest,
		copy.Options{
			OnSymlink:   func(string) copy.SymlinkAction { return copy.Shallow },
			OnDirExists: func(string, string) copy.DirExistsAction { return copy.Merge },
			// OnError: func(_src, _dest string, e error) error {
			// 	return e
			// },
		})
}
