package gox

import (
	"fmt"
	"io"
	"os"
	"path"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"

	"golang.org/x/mod/modfile"
)

type GoModInfo struct {
	ModPath   string
	Toolchain string
}

// GetModulePath returns the module path and the toolchain version
// of the `go.mod` file located in `rootDir`.
// The toolchain string might be not set.
func GetModuleInfo(rootDir string) (info GoModInfo, err error) {
	goMod := path.Join(rootDir, "go.mod")
	file, err := os.OpenFile(goMod, os.O_RDONLY, fs.DefaultPermissionsFile)
	if err != nil {
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	m, err := modfile.Parse(goMod, data, nil)
	if err != nil {
		return
	}

	if m.Module != nil {
		info.ModPath = m.Module.Mod.Path
	} else {
		err = fmt.Errorf("no module path in 'go.mod' file '%v'", goMod)

		return
	}

	if m.Toolchain != nil {
		info.Toolchain = m.Toolchain.Name
	}

	return
}
