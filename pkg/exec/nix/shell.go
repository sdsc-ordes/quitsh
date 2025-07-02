package nix

import (
	"os"
	"slices"
	"strings"
)

// The toolchain env. variable as a comma separated
// list of toolchain names in the current shell.
const EnvVarQuitshToolchain = "QUITSH_TOOLCHAINS"

// Enable `--no-pure-eval` on Nix DevShells, to forexample pass
// env. values which are evaluated in Nix.
const EnvVarQuitshNixNoPureEval = "QUITSH_NIX_NO_PURE_EVAL"

// InBuild returns `true` if we are inside a Nix build.
func InBuild() bool {
	_, inBuild := os.LookupEnv("NIX_BUILD_TOP")

	return inBuild
}

// InShell returns `true` if we are inside a Nix shell.
func InShell() bool {
	_, inShell := os.LookupEnv("IN_NIX_SHELL")

	return inShell
}

// HardeningOptions returns the set hardening options.
func HardeningOptions() []string {
	return strings.Split(os.Getenv("NIX_HARDENING_ENABLE"), " ")
}

// HaveToolchain tests if we are running inside a
// Nix shell/Nix build
// which has the toolchain `toolchain` available.
// A toolchain is just a set of tools we define.
// The Nix shells need to have a `QUITSH_TOOLCHAINS = "a,b,c"`
// set.
func HaveToolchain(toolchain string) bool {
	if toolchain == "nix" {
		// The toolchain "nix" is always available.
		// `nix` is a prerequisite.
		return true
	}

	if InBuild() || InShell() {
		tcs := os.Getenv(EnvVarQuitshToolchain)
		toolchains := strings.Split(strings.ReplaceAll(tcs, " ", ""), ",")

		return slices.Contains(toolchains, toolchain)
	}

	return false
}

// ToolchainInstallable returns the `nix develop <toolchainref>` string
// for the given toolchain.
func ToolchainInstallable(flakePath string, toolchain string) string {
	return FlakeInstallable(flakePath, toolchain)
}
