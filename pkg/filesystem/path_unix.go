//go:build unix

package fs

import "golang.org/x/sys/unix"

// IsExecutable checks if the `path` is executable.
func IsExecutable(path string) bool {
	return unix.Access(path, unix.X_OK) == nil
}
