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
) (packages map[string]*Package, err error) {
	pkgs, err := getFlakeAttributes(rootDir, flakePath, "packages")
	if err != nil {
		return nil, err
	}

	packages = make(map[string]*Package, len(pkgs))
	for _, p := range pkgs {
		packages[p.Name] = p
	}

	return packages, nil
}

// GetFlakeShells is similar to [GetFlakeShells] but reports all Dev shells.
func GetFlakeShells(
	rootDir string,
	flakePath string,
) (packages map[string]*Package, err error) {
	pkgs, err := getFlakeAttributes(rootDir, flakePath, "devShells")
	if err != nil {
		return nil, err
	}

	packages = make(map[string]*Package, len(pkgs))
	for _, p := range pkgs {
		packages[p.Name] = p
	}

	return packages, nil
}

// GetFlakeOutputs returns all flake outputs for current system
// such as `<prefix>.<currentSystem>.X`.
func GetFlakeOutputs(
	rootDir string,
	flakePath string,
	prefixes []string,
) (packages []*Package, err error) {
	for _, p := range prefixes {
		m, e := getFlakeAttributes(rootDir, flakePath, p)
		if e != nil {
			return nil, e
		}

		packages = append(packages, m...)
	}

	return packages, nil
}

func getFlakeAttributes(
	rootDir string,
	flakePath string,
	prefixPath string,
) (packages []*Package, err error) {
	nixx := NewEvalCtx(rootDir)

	currentSystem, err := CurrentSystem()
	if err != nil {
		return nil, err
	}

	attrPath := prefixPath + "." + currentSystem

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

	packages = make([]*Package, 0, len(json))
	for name, storePath := range json {
		packages = append(
			packages,
			&Package{Name: name, StorePath: storePath, AttrPath: attrPath + "." + name},
		)
	}

	return packages, nil
}
