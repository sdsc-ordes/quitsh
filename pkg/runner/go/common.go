package gorunner

import (
	"path"
	"slices"
	"strings"

	cm "github.com/sdsc-ordes/quitsh/pkg/common"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/hashicorp/go-version"
)

func GetBuildFlags(
	log log.ILog,
	compDir string,
	buildType cm.BuildType,
	envType cm.EnvironmentType,
	coverage bool,
	verbose bool,
	modInfo gox.GoModInfo,
	version *version.Version,
	versionModule string,
	buildTags []string,
	isTest bool,
) (flags []string) {
	flags = []string{"-C", compDir}

	bTags := append([]string{}, buildTags...)
	var ldFlags []string

	if buildType == cm.BuildDebug {
		log.Warn("Building debug version.")
		bTags = append(bTags, "debug")
	} else {
		// Strip debug information (-w).
		// Strip symbol table (-s).
		ldFlags = append(ldFlags, "-s", "-w")
	}

	// Set the <versionModule>.BuildVersion string
	if versionModule == "" {
		versionModule = "build"
	}

	ldFlags = append(
		ldFlags,
		"-X", path.Join(
			modInfo.ModPath,
			versionModule+
				".buildVersion=",
		)+version.String(),
	)

	// Nix build options.
	if nix.InBuild() {
		if !isTest {
			// When you build a Go program, the paths to source
			// files are embedded in the resulting binary for debugging and stack traces.
			// These paths usually include full file system paths,
			// which means if you build the same code in different
			// directories or on different machines, the binaries will differ.
			flags = append(flags, "--trimpath")
		}

		// Set the buildid empty (Go stores it in the binary) for reproducibility.
		ldFlags = append(ldFlags, "--buildid=")

		// Set the hardening options if enabled.
		if slices.Contains(nix.HardeningOptions(), "pie") {
			flags = append(flags, "--buildmode=pie")
		}
	}

	if coverage {
		bTags = append(bTags, "coverage")
		flags = append(flags,
			"--cover",
			"--covermode=atomic",
		)
	}

	// Add tags.
	// Add default `test` flag when testing.
	if isTest {
		bTags = append(bTags, "test")
	}

	// Append the environment tag.
	bTags = append(bTags, envType.String())

	if len(bTags) != 0 {
		flags = append(flags, "--tags", strings.Join(bTags, ","))
	}
	if len(ldFlags) != 0 {
		flags = append(flags, "--ldflags", strings.Join(ldFlags, " "))
	}

	// Verbose mode.
	if verbose {
		flags = append(flags, "-v")
	}

	log.Info("Build flags.", "flags", flags)

	return flags
}
