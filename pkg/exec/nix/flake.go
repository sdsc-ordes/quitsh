package nix

import (
	"encoding/json"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

type Package struct {
	// The package name as usable in `nix build ./dir-to-flake#<name>`.
	Name string

	// The full package attribute path, e.g. `packages.<system>.<name>`.
	// with the current system.
	AttrPath string

	// The package store path.
	StorePath string
}

// GetFlakePackages gets all packages from the flake for the current system.
// This uses `nix eval <flakeOutputAttr>`
// Note: We cannot use `nix flake show` (sadly) because we use IFD (import from derivation inside `yaml.nix`)
// which is unfortunate (maybe we can later somehow convert the `.component.yaml`s to JSON)
// See also: https://github.com/NixOS/nix/issues/4265
func GetFlakePackages(
	rootDir string,
	flakePath string,
) (packages map[string]Package, err error) {
	nixx := NewEvalCtx(rootDir)

	currentSystem, err := CurrentSystem()
	if err != nil {
		return nil, err
	}

	attrPath := "packages." + currentSystem

	cmd := []string{"--json", FlakeInstallable(flakePath, attrPath)}

	jsonRes, err := nixx.Get(cmd...)
	if err != nil {
		return nil, errors.AddContext(err, "could not evaluate package in flake.")
	}

	decoder := json.NewDecoder(strings.NewReader(jsonRes))
	json := map[string]string{}

	err = decoder.Decode(&json)
	if err != nil {
		return nil, errors.AddContext(err, "Could not decode Nix json result.")
	}

	packages = make(map[string]Package, len(json))
	for name, storePath := range json {
		packages[name] = Package{Name: name, StorePath: storePath, AttrPath: attrPath + "." + name}
	}

	return
}
