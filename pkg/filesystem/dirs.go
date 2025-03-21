package fs

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// MkDirs creates all directories in `path`.
func MkDirs(paths ...string) (err error) {
	for i := range paths {
		err = errors.Combine(err,
			os.MkdirAll(paths[i], DefaultPermissionsDir))
	}

	return
}

// AssertDirs asserts that `path` exists, otherwise panics.
func AssertDirs(paths ...string) {
	err := MkDirs(paths...)
	log.PanicE(err, "Could not create dir.", "paths", paths)
}
