package nix

import (
	"fmt"
	"strings"
)

// FlakeInstallable returns the attribute path `<flakePath>#<attrPath>`.
func FlakeInstallable(flakePath string, attrPath string) string {
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
