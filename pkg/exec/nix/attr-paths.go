package nix

import (
	"fmt"
	"path"
	"strings"
)

// FlakeInstallable returns the attribute path `<flakePath>#<attrPath>`.
func FlakeInstallable(flakePath string, attrPath string) string {
	if !path.IsAbs(flakePath) && !strings.HasPrefix(flakePath, ".") {
		// We need always a `./...#` because `test/bla#package` does not work.
		// NOTE: path.Join cleans the result.
		flakePath = "./" + flakePath
	}

	return fmt.Sprintf("%s#%s", flakePath, attrPath)
}

// ReplaceSystem replaces placeholder "${system}" in string `attrPath` with
// the current system.
func ReplaceSystem(attrPath string) (string, error) {
	system, err := CurrentSystem()
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(attrPath, "${system}", system), nil
}
