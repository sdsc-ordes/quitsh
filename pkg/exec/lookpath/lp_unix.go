//go:build unix

// Taken from: https://cs.opensource.google/go/go/+/refs/tags/go1.21.6:src/os/exec/lp_unix.go;l=52
// to adapt LookPath for our needs not consulting `Getenv("PATH")`.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//nolint:nlreturn,errorlint // External file from Stdlib.
package lookpath

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// Error is returned by [LookPath] when it fails to classify a file as an
// executable.
type Error struct {
	// Name is the file name for which the error occurred.
	Name string
	// Err is the underlying error.
	Err error
}

func (e *Error) Error() string {
	return "exec: " + strconv.Quote(e.Name) + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error { return e.Err }

// ErrDot indicates that a path lookup resolved to an executable
// in the current directory due to ‘.’ being in the path, either
// implicitly or explicitly. See the package documentation for details.
//
// Note that functions in this package do not return ErrDot directly.
// Code should use errors.Is(err, ErrDot), not err == ErrDot,
// to test whether a returned error err is due to this condition.
var ErrDot = errors.New("cannot run executable found relative to current directory")

// ErrNotFound is the error resulting if a path search failed to find an executable file.
var ErrNotFound = errors.New("executable file not found in $PATH")

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}

	m := d.Mode()
	if m.IsDir() {
		return syscall.EISDIR
	}

	// NOTE: We need to replace it with the toplevel
	//       Access function (darwin does not enable EAccess to wrap).
	//       This wraps call https://cs.opensource.google/go/go/+/refs/tags/go1.21.6:src/os/exec/lp_unix.go;l=31
	err = unix.Access(file, unix.X_OK)

	// ENOSYS means Eaccess is not available or not implemented.
	// EPERM can be returned by Linux containers employing seccomp.
	// In both cases, fall back to checking the permission bits.
	if err == nil || (err != syscall.ENOSYS && err != syscall.EPERM) {
		return err
	}
	if m&0111 != 0 {
		return nil
	}
	return fs.ErrPermission
}

// LookPath searches for an executable named file in the
// directories named by the `path` (env. variable).
// If file contains a slash, it is tried directly and the PATH is not consulted.
// Otherwise, on success, the result is an absolute path.
// See the https://cs.opensource.google/go/go/+/refs/tags/go1.21.6:src/os/exec/lp_unix.go;l=52
// for further details.
func Look(file string, pathEnv string) (string, error) {
	// NOTE(rsc): I wish we could use the Plan 9 behavior here
	// (only bypass the path if file begins with / or ./ or ../)
	// but that would not match all the Unix shells.

	if strings.Contains(file, "/") {
		err := findExecutable(file)
		if err == nil {
			return file, nil
		}

		return "", &Error{file, err}
	}

	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		err := findExecutable(path)
		if err == nil {
			if !filepath.IsAbs(path) {
				return path, &Error{file, ErrDot}
			}
			return path, nil
		}
	}

	return "", &Error{file, ErrNotFound}
}
