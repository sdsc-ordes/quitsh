package nix

import "path"

func GetNixFlakeDir(rootDir string) string {
	return path.Join(rootDir, "tools/nix")
}
